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

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	kaitov1beta1 "github.com/kaito-project/kaito/api/v1beta1"
	"github.com/kaito-project/kaito/pkg/nodeprovision"
	"github.com/kaito-project/kaito/pkg/utils/consts"
	"github.com/kaito-project/kaito/pkg/utils/nodeclaim"
)

const (
	// NodeClassNotReadyReason reports a missing or Ready=False NodeClass, which blocks every launch.
	NodeClassNotReadyReason = "NodeClassNotReady"
	// nodePoolValidationFailedReason is the NodePool ValidationSucceeded=False reason written by Karpenter.
	nodePoolValidationFailedReason = "NodePoolValidationFailed"
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
	// genericInsufficientCapacityMessages are the insufficient capacity errors of a NodeClaim whose instance types have
	// no available offering. KAITO NodePools allow a single instance type and capacity type, so its only offering is
	// cached as unavailable after a quota, allocation or SKU failure, or not offered at all.
	genericInsufficientCapacityMessages = []string{"all requested instance types were unavailable",
		"no instance types available"}
	nodeImageInvalidCodes = sets.New("ImageNotFound", "GalleryImageNotFound")
	// nodeConfigurationInvalidReasons are checked on every reconcile, so a failure they reported is never kept once fixed.
	nodeConfigurationInvalidReasons = sets.New(NodeClassNotReadyReason, nodePoolValidationFailedReason)
	nodeConfigurationInvalidCodes   = sets.New("AuthorizationFailed", "LinkedAuthorizationFailed", "SubnetIsFull",
		"InsufficientSubnetSize")
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
		containsAny(lower, capacityUnavailableMessages...) || isGenericInsufficientCapacity(reason, message):
		return consts.ReasonCapacityUnavailable
	case codes.HasAny(nodeImageInvalidCodes.UnsortedList()...) ||
		codes.Has("InvalidParameter") && strings.Contains(lower, "imagereference"):
		return consts.ReasonNodeImageInvalid
	case reason == registrationFailedReason || codes.Has("VMExtensionProvisioningError"):
		return consts.ReasonNodeBootstrapFailed
	case nodeConfigurationInvalidReasons.Has(reason) || codes.HasAny(nodeConfigurationInvalidCodes.UnsortedList()...):
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

// isGenericInsufficientCapacity reports an insufficient capacity error that does not say why the instance type is
// unavailable, as it follows an earlier failure that the cloud provider cached.
func isGenericInsufficientCapacity(reason, message string) bool {
	return reason == insufficientCapacityEventReason && containsAny(strings.ToLower(message), genericInsufficientCapacityMessages...)
}

// previousSpecificCause returns the Workspace's current NodeClaimReady failure for its generation unless it is
// progress or an invalid NodeClass or NodePool, which is reported only while it lasts. generic reports a failure
// classified from a generic insufficient capacity error.
func previousSpecificCause(ws *kaitov1beta1.Workspace) (reason, message string, generic, found bool) {
	c := meta.FindStatusCondition(ws.Status.Conditions, string(kaitov1beta1.ConditionTypeNodeClaimStatus))
	if c == nil || c.Status != metav1.ConditionFalse || c.ObservedGeneration != ws.Generation || c.Reason == "" ||
		c.Reason == nodeclaim.NodeClaimNotReadyReason || c.Reason == insufficientCapacityEventReason {
		return "", "", false, false
	}
	message, _, _ = strings.Cut(c.Message, nodeprovision.ReconcileErrorMessagePrefix)
	// A classified message starts with the original reason.
	originalReason, originalMessage, _ := strings.Cut(message, ": ")
	if nodeConfigurationInvalidReasons.Has(originalReason) {
		return "", "", false, false
	}
	return c.Reason, message, isGenericInsufficientCapacity(originalReason, originalMessage), true
}
