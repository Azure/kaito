// Copyright (c) KAITO authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package karpenter

import (
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"

	kaitov1beta1 "github.com/kaito-project/kaito/api/v1beta1"
	"github.com/kaito-project/kaito/pkg/nodeprovision"
	"github.com/kaito-project/kaito/pkg/utils/consts"
	"github.com/kaito-project/kaito/pkg/utils/nodeclaim"
)

const (
	// registrationFailedReason is the NodePool NodeRegistrationHealthy=False reason written by Karpenter when launched
	// nodes repeatedly fail to register.
	registrationFailedReason = "RegistrationFailed"
	// subscriptionQuotaReachedReason is the Launched reason of the Azure provider for an exhausted vCPU quota.
	subscriptionQuotaReachedReason = "SubscriptionQuotaReached"
)

var (
	capacityUnavailableReasons = sets.New("AllocationFailure", "ZonalAllocationFailure",
		"OverconstrainedAllocationFailure", "OverconstrainedZonalAllocationFailure", "SKUNotAvailable")
	capacityUnavailableCodes = sets.New("AllocationFailed", "ZonalAllocationFailed",
		"OverconstrainedAllocationRequest", "OverconstrainedZonalAllocationRequest", "SkuNotAvailable", "VMSizeNotSupported")
	// capacityUnavailableMessages are written by the Azure provider's allocation and SKU error handlers.
	capacityUnavailableMessages = []string{"unable to allocate resources", "the requested sku is unavailable",
		"is not supported for this subscription"}
	// noOfferingMessages are the insufficient capacity errors of a NodeClaim whose instance types have no available
	// offering, which do not say why. KAITO NodePools allow a single instance type and capacity type, so its only
	// offering is either cached as unavailable after an earlier quota, allocation or SKU failure, whose NodeClaim is
	// deleted, or not offered at all.
	noOfferingMessages            = []string{"all requested instance types were unavailable", "no instance types available"}
	nodeImageInvalidCodes         = sets.New("ImageNotFound", "GalleryImageNotFound")
	nodeConfigurationInvalidCodes = sets.New("AuthorizationFailed", "LinkedAuthorizationFailed", "SubnetIsFull",
		"InsufficientSubnetSize")
	// nodePoolOrNodeClassReasons report an invalid NodePool or a failed NodeClass, read directly on every reconcile.
	nodePoolOrNodeClassReasons = sets.New(consts.ReasonNodeClassInvalid, consts.ReasonNodePoolInvalid)
)

func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// classifyProvisioningFailure returns the KAITO reason of a provisioning failure that needs the user to act, or ""
// otherwise. It matches the Launched reasons of recent Azure providers, the messages of their error handlers, which
// older providers report as CreateInstanceFailed or in Events, and the cloud error codes.
func classifyProvisioningFailure(reason, message string) string {
	codes := nodeclaim.CloudErrorCodes(message)
	lower := strings.ToLower(message)
	switch {
	case reason == subscriptionQuotaReachedReason || codes.Has("QuotaExceeded") ||
		containsAny(lower, "vcpu quota", "lowpriorityquota") ||
		codes.Has("OperationNotAllowed") && strings.Contains(lower, "quota"):
		return consts.ReasonQuotaExceeded
	case capacityUnavailableReasons.Has(reason) || codes.HasAny(capacityUnavailableCodes.UnsortedList()...) ||
		containsAny(lower, capacityUnavailableMessages...) || isNoOfferingAvailable(reason, message):
		return consts.ReasonCapacityUnavailable
	case codes.HasAny(nodeImageInvalidCodes.UnsortedList()...) ||
		codes.Has("InvalidParameter") && strings.Contains(lower, "imagereference"):
		return consts.ReasonNodeImageInvalid
	case reason == registrationFailedReason || codes.Has("VMExtensionProvisioningError"):
		return consts.ReasonNodeBootstrapFailed
	case codes.HasAny(nodeConfigurationInvalidCodes.UnsortedList()...):
		return consts.ReasonNodeConfigurationInvalid
	}
	return ""
}

// provisioningCause returns the reason and message reported for a provisioning failure: a failure that needs the user
// to act is reported under its KAITO reason, with the original reason in front of the message.
func provisioningCause(reason, message string) (string, string) {
	if classified := classifyProvisioningFailure(reason, message); classified != "" {
		return classified, nodeclaim.FormatClassifiedProvisioningMessage(reason, message)
	}
	return nodeclaim.ValidConditionReason(reason), nodeclaim.FormatProvisioningMessage(message)
}

// isNoOfferingAvailable reports an insufficient capacity error that no instance type offering is available, which
// does not say why.
func isNoOfferingAvailable(reason, message string) bool {
	return reason == insufficientCapacityEventReason && containsAny(strings.ToLower(message), noOfferingMessages...)
}

// reportedNotReadyCause is the NodeClaimReady failure already reported on the Workspace.
type reportedNotReadyCause struct {
	reason  string
	message string
	// noOffering reports a failure classified from a no-offering error.
	noOffering bool
}

// reportedCause returns the NodeClaimReady failure already reported on the Workspace while it still applies. The
// NodeClaim that carried a failure is deleted, and its retries often fail before launching with an error that does not
// say why, so the failure is kept on the Workspace until a new attempt launches.
func reportedCause(ws *kaitov1beta1.Workspace, nodeClaims []*karpenterv1.NodeClaim, now time.Time) (reportedNotReadyCause, bool) {
	c := meta.FindStatusCondition(ws.Status.Conditions, string(kaitov1beta1.ConditionTypeNodeClaimStatus))
	if c == nil || c.Status != metav1.ConditionFalse || c.ObservedGeneration != ws.Generation {
		return reportedNotReadyCause{}, false
	}
	// Only a failure that needs the user to act is kept: progress, or a failure KAITO does not classify, says
	// nothing once its NodeClaim is gone.
	if !consts.IsTerminatedReason(c.Reason) {
		return reportedNotReadyCause{}, false
	}
	// The NodePool and NodeClass are read directly on every reconcile, so a failure of theirs is never kept once fixed.
	if nodePoolOrNodeClassReasons.Has(c.Reason) {
		return reportedNotReadyCause{}, false
	}
	message, _, _ := strings.Cut(c.Message, nodeprovision.ReconcileErrorMessagePrefix)
	originalReason, originalMessage := nodeclaim.ParseClassifiedProvisioningMessage(message)
	// The failure of the earlier attempts no longer applies once a new attempt launched.
	if hasUnregisteredLaunch(nodeClaims, now, 0) {
		return reportedNotReadyCause{}, false
	}
	return reportedNotReadyCause{
		reason:     c.Reason,
		message:    message,
		noOffering: isNoOfferingAvailable(originalReason, originalMessage),
	}, true
}
