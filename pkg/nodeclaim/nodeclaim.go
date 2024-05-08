// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package nodeclaim

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

const (
	KaitoNodePoolName             = "kaito"
	LabelNodePool                 = "karpenter.sh/nodepool"
	ErrorInstanceTypesUnavailable = "all requested instance types were unavailable during launch"
)

var (
	// nodeClaimStatusTimeoutInterval is the interval to check the nodeClaim status.
	nodeClaimStatusTimeoutInterval = 240 * time.Second
)

// GenerateNodeClaimManifest generates a nodeClaim object from the given workspace.
func GenerateNodeClaimManifest(ctx context.Context, storageRequirement string, workspaceObj *kaitov1alpha1.Workspace) *v1beta1.NodeClaim {
	digest := sha256.Sum256([]byte(workspaceObj.Namespace + workspaceObj.Name + time.Now().
		Format("2006-01-02 15:04:05.000000000"))) // We make sure the nodeClaim name is not fixed to the workspace
	nodeClaimName := "ws" + hex.EncodeToString(digest[0:])[0:9]

	nodeClaimLabels := map[string]string{
		LabelNodePool:                         KaitoNodePoolName, // Fake nodepool name to prevent Karpenter from scaling up.
		kaitov1alpha1.LabelWorkspaceName:      workspaceObj.Name,
		kaitov1alpha1.LabelWorkspaceNamespace: workspaceObj.Namespace,
		v1beta1.DoNotDisruptAnnotationKey:     "true", // To prevent Karpenter from scaling down.
	}
	if workspaceObj.Resource.LabelSelector != nil &&
		len(workspaceObj.Resource.LabelSelector.MatchLabels) != 0 {
		nodeClaimLabels = lo.Assign(nodeClaimLabels, workspaceObj.Resource.LabelSelector.MatchLabels)
	}

	return &v1beta1.NodeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeClaimName,
			Namespace: workspaceObj.Namespace,
			Labels:    nodeClaimLabels,
		},
		Spec: v1beta1.NodeClaimSpec{
			Requirements: []v1beta1.NodeSelectorRequirementWithMinValues{
				{
					NodeSelectorRequirement: v1.NodeSelectorRequirement{
						Key:      v1.LabelInstanceTypeStable,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{workspaceObj.Resource.InstanceType},
					},
					MinValues: lo.ToPtr(1),
				},
				{
					NodeSelectorRequirement: v1.NodeSelectorRequirement{
						Key:      LabelNodePool,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{KaitoNodePoolName},
					},
					MinValues: lo.ToPtr(1),
				},
				{
					NodeSelectorRequirement: v1.NodeSelectorRequirement{
						Key:      v1.LabelArchStable,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"amd64"},
					},
					MinValues: lo.ToPtr(1),
				},
				{
					NodeSelectorRequirement: v1.NodeSelectorRequirement{
						Key:      v1.LabelOSStable,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"linux"},
					},
					MinValues: lo.ToPtr(1),
				},
			},
			Taints: []v1.Taint{
				{
					Key:    "sku",
					Value:  "gpu",
					Effect: v1.TaintEffectNoSchedule,
				},
			},
			Resources: v1beta1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(storageRequirement),
				},
			},
		},
	}
}

// CreateNodeClaim creates a nodeClaim object.
func CreateNodeClaim(ctx context.Context, nodeClaimObj *v1beta1.NodeClaim, kubeClient client.Client) error {
	klog.InfoS("CreateNodeClaim", "nodeClaim", klog.KObj(nodeClaimObj))
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return err.Error() != ErrorInstanceTypesUnavailable
	}, func() error {
		err := kubeClient.Create(ctx, nodeClaimObj, &client.CreateOptions{})
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)

		updatedObj := &v1beta1.NodeClaim{}
		err = kubeClient.Get(ctx, client.ObjectKey{Name: nodeClaimObj.Name, Namespace: nodeClaimObj.Namespace}, updatedObj, &client.GetOptions{})

		// if SKU is not available, then exit.
		_, conditionFound := lo.Find(updatedObj.GetConditions(), func(condition apis.Condition) bool {
			return condition.Type == v1beta1.Launched &&
				condition.Status == v1.ConditionFalse && condition.Message == ErrorInstanceTypesUnavailable
		})
		if conditionFound {
			klog.Error(ErrorInstanceTypesUnavailable, "reconcile will not continue")
			return fmt.Errorf(ErrorInstanceTypesUnavailable)
		}
		return err
	})
}

// WaitForPendingNodeClaims checks if the there are any nodeClaims in provisioning condition. If so, wait until they are ready.
func WaitForPendingNodeClaims(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, kubeClient client.Client) error {
	nodeClaims, err := ListNodeClaimByWorkspace(ctx, workspaceObj, kubeClient)
	if err != nil {
		return err
	}

	for i := range nodeClaims.Items {
		// check if the nodeClaim is being created has the requested workspace instance type.
		_, nodeClaimInstanceType := lo.Find(nodeClaims.Items[i].Spec.Requirements, func(requirement v1beta1.NodeSelectorRequirementWithMinValues) bool {
			return requirement.Key == v1.LabelInstanceTypeStable &&
				requirement.Operator == v1.NodeSelectorOpIn &&
				lo.Contains(requirement.Values, workspaceObj.Resource.InstanceType)
		})
		if nodeClaimInstanceType {
			_, found := lo.Find(nodeClaims.Items[i].GetConditions(), func(condition apis.Condition) bool {
				return condition.Type == v1beta1.Initialized && condition.Status == v1.ConditionFalse
			})

			if found || nodeClaims.Items[i].GetConditions() == nil { // checking conditions==nil is a workaround for conditions delaying to set on the nodeClaim object.
				//wait until nodeClaim is initialized.
				if err := CheckNodeClaimStatus(ctx, &nodeClaims.Items[i], kubeClient); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ListNodeClaimByWorkspace list all nodeClaim objects in the cluster that are created by the workspace identified by the label.
func ListNodeClaimByWorkspace(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, kubeClient client.Client) (*v1beta1.NodeClaimList, error) {
	nodeClaimList := &v1beta1.NodeClaimList{}

	ls := labels.Set{
		kaitov1alpha1.LabelWorkspaceName:      workspaceObj.Name,
		kaitov1alpha1.LabelWorkspaceNamespace: workspaceObj.Namespace,
	}

	err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.List(ctx, nodeClaimList, &client.MatchingLabelsSelector{Selector: ls.AsSelector()})
	})
	if err != nil {
		return nil, err
	}

	return nodeClaimList, nil
}

// CheckNodeClaimStatus checks the status of the nodeClaim. If the nodeClaim is not ready, then it will wait for the nodeClaim to be ready.
// If the nodeClaim is not ready after the timeout, then it will return an error.
// if the nodeClaim is ready, then it will return nil.
func CheckNodeClaimStatus(ctx context.Context, nodeClaimObj *v1beta1.NodeClaim, kubeClient client.Client) error {
	klog.InfoS("CheckNodeClaimStatus", "nodeClaim", klog.KObj(nodeClaimObj))
	timeClock := clock.RealClock{}
	tick := timeClock.NewTicker(nodeClaimStatusTimeoutInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-tick.C():
			return fmt.Errorf("check nodeClaim status timed out. nodeClaim %s is not ready", nodeClaimObj.Name)

		default:
			time.Sleep(1 * time.Second)
			err := kubeClient.Get(ctx, client.ObjectKey{Name: nodeClaimObj.Name, Namespace: nodeClaimObj.Namespace}, nodeClaimObj, &client.GetOptions{})
			if err != nil {
				return err
			}

			// if nodeClaim is not ready, then continue.
			_, conditionFound := lo.Find(nodeClaimObj.GetConditions(), func(condition apis.Condition) bool {
				return condition.Type == apis.ConditionReady &&
					condition.Status == v1.ConditionTrue
			})
			if !conditionFound {
				continue
			}

			klog.InfoS("nodeClaim status is ready", "nodeClaim", nodeClaimObj.Name)
			return nil
		}
	}
}
