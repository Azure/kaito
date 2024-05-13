// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package machine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
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
)

const (
	ProvisionerName               = "default"
	LabelGPUProvisionerCustom     = "kaito.sh/machine-type"
	LabelProvisionerName          = "karpenter.sh/provisioner-name"
	GPUString                     = "gpu"
	ErrorInstanceTypesUnavailable = "all requested instance types were unavailable during launch"
)

var (
	// machineStatusTimeoutInterval is the interval to check the machine status.
	machineStatusTimeoutInterval = 240 * time.Second
)

// GenerateMachineManifest generates a machine object from the given workspace.
func GenerateMachineManifest(ctx context.Context, storageRequirement string, workspaceObj *kaitov1alpha1.Workspace) *v1alpha5.Machine {
	digest := sha256.Sum256([]byte(workspaceObj.Namespace + workspaceObj.Name + time.Now().Format("2006-01-02 15:04:05.000000000"))) // We make sure the machine name is not fixed to the a workspace
	machineName := "ws" + hex.EncodeToString(digest[0:])[0:9]
	machineLabels := map[string]string{
		LabelProvisionerName:                  ProvisionerName,
		kaitov1alpha1.LabelWorkspaceName:      workspaceObj.Name,
		kaitov1alpha1.LabelWorkspaceNamespace: workspaceObj.Namespace,
	}
	if workspaceObj.Resource.LabelSelector != nil &&
		len(workspaceObj.Resource.LabelSelector.MatchLabels) != 0 {
		machineLabels = lo.Assign(machineLabels, workspaceObj.Resource.LabelSelector.MatchLabels)

	}

	return &v1alpha5.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineName,
			Namespace: workspaceObj.Namespace,
			Labels:    machineLabels,
		},
		Spec: v1alpha5.MachineSpec{
			MachineTemplateRef: &v1alpha5.MachineTemplateRef{
				Name: machineName,
			},
			Requirements: []v1.NodeSelectorRequirement{
				{
					Key:      v1.LabelInstanceTypeStable,
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{workspaceObj.Resource.InstanceType},
				},
				{
					Key:      LabelProvisionerName,
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{ProvisionerName},
				},
				{
					Key:      LabelGPUProvisionerCustom,
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{GPUString},
				},
				{
					Key:      v1.LabelArchStable,
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{"amd64"},
				},
				{
					Key:      v1.LabelOSStable,
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{"linux"},
				},
			},
			Taints: []v1.Taint{
				{
					Key:    "sku",
					Value:  GPUString,
					Effect: v1.TaintEffectNoSchedule,
				},
			},
			Resources: v1alpha5.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(storageRequirement),
				},
			},
		},
	}
}

// CreateMachine creates a machine object.
func CreateMachine(ctx context.Context, machineObj *v1alpha5.Machine, kubeClient client.Client) error {
	klog.InfoS("CreateMachine", "machine", klog.KObj(machineObj))
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return err.Error() != ErrorInstanceTypesUnavailable
	}, func() error {
		err := kubeClient.Create(ctx, machineObj, &client.CreateOptions{})
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)

		updatedObj := &v1alpha5.Machine{}
		err = kubeClient.Get(ctx, client.ObjectKey{Name: machineObj.Name, Namespace: machineObj.Namespace}, updatedObj, &client.GetOptions{})

		// if SKU is not available, then exit.
		_, conditionFound := lo.Find(updatedObj.GetConditions(), func(condition apis.Condition) bool {
			return condition.Type == v1alpha5.MachineLaunched &&
				condition.Status == v1.ConditionFalse && condition.Message == ErrorInstanceTypesUnavailable
		})
		if conditionFound {
			klog.Error(ErrorInstanceTypesUnavailable, "reconcile will not continue")
			return fmt.Errorf(ErrorInstanceTypesUnavailable)
		}
		return err
	})
}

// WaitForPendingMachines checks if the there are any machines in provisioning condition. If so, wait until they are ready.
func WaitForPendingMachines(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, kubeClient client.Client) error {
	machines, err := ListMachinesByWorkspace(ctx, workspaceObj, kubeClient)
	if err != nil {
		return err
	}

	for i := range machines.Items {
		// check if the machine is being created has the requested workspace instance type.
		_, machineInstanceType := lo.Find(machines.Items[i].Spec.Requirements, func(requirement v1.NodeSelectorRequirement) bool {
			return requirement.Key == v1.LabelInstanceTypeStable &&
				requirement.Operator == v1.NodeSelectorOpIn &&
				lo.Contains(requirement.Values, workspaceObj.Resource.InstanceType)
		})
		if machineInstanceType {
			_, found := lo.Find(machines.Items[i].GetConditions(), func(condition apis.Condition) bool {
				return condition.Type == v1alpha5.MachineInitialized && condition.Status == v1.ConditionFalse
			})

			if found || machines.Items[i].GetConditions() == nil { // checking conditions==nil is a workaround for conditions delaying to set on the machine object.
				//wait until machine is initialized.
				if err := CheckMachineStatus(ctx, &machines.Items[i], kubeClient); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ListMachinesByWorkspace list all machine objects in the cluster that are created by the workspace identified by the label.
func ListMachinesByWorkspace(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, kubeClient client.Client) (*v1alpha5.MachineList, error) {
	machineList := &v1alpha5.MachineList{}

	ls := labels.Set{
		kaitov1alpha1.LabelWorkspaceName:      workspaceObj.Name,
		kaitov1alpha1.LabelWorkspaceNamespace: workspaceObj.Namespace,
	}

	err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.List(ctx, machineList, &client.MatchingLabelsSelector{Selector: ls.AsSelector()})
	})
	if err != nil {
		return nil, err
	}

	return machineList, nil
}

// CheckMachineStatus checks the status of the machine. If the machine is not ready, then it will wait for the machine to be ready.
// If the machine is not ready after the timeout, then it will return an error.
// if the machine is ready, then it will return nil.
func CheckMachineStatus(ctx context.Context, machineObj *v1alpha5.Machine, kubeClient client.Client) error {
	klog.InfoS("CheckMachineStatus", "machine", klog.KObj(machineObj))
	timeClock := clock.RealClock{}
	tick := timeClock.NewTicker(machineStatusTimeoutInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-tick.C():
			return fmt.Errorf("check machine status timed out. machine %s is not ready", machineObj.Name)

		default:
			time.Sleep(1 * time.Second)
			err := kubeClient.Get(ctx, client.ObjectKey{Name: machineObj.Name, Namespace: machineObj.Namespace}, machineObj, &client.GetOptions{})
			if err != nil {
				return err
			}

			// if machine is not ready, then continue.
			_, conditionFound := lo.Find(machineObj.GetConditions(), func(condition apis.Condition) bool {
				return condition.Type == apis.ConditionReady &&
					condition.Status == v1.ConditionTrue
			})
			if !conditionFound {
				continue
			}

			klog.InfoS("machine status is ready", "machine", machineObj.Name)
			return nil
		}
	}
}
