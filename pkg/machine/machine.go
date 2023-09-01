package machine

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/kdm/api/v1alpha1"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProvisionerName               = "default"
	LabelGPUProvisionerCustom     = "gpu-provisioner.sh/machine-type"
	LabelProvisionerName          = "karpenter.sh/provisioner-name"
	GPUString                     = "gpu"
	ErrorInstanceTypesUnavailable = "all requested instance types were unavailable during launch"
)

var (
	machineStatusCheckInterval                  = 60 * time.Second
	timeClock                  clock.WithTicker = clock.RealClock{}
)

// GenerateMachineManifest generates a machine object from	the given workspace.
func GenerateMachineManifest(ctx context.Context, workspaceObj *v1alpha1.Workspace) *v1alpha5.Machine {
	klog.InfoS("GenerateMachineManifest", "workspace", klog.KObj(workspaceObj))

	machineName := fmt.Sprint("machine", rand.Intn(100_000))

	machineLabels := map[string]string{
		LabelProvisionerName: ProvisionerName,
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

// CheckMachineStatus checks the status of the machine. If the machine is not ready, then it will wait for the machine to be ready.
// If the machine is not ready after the timeout, then it will return an error.
// if	the machine is ready, then it will return nil.
func CheckMachineStatus(ctx context.Context, machineObj *v1alpha5.Machine, kubeClient client.Client) error {
	klog.InfoS("CheckMachineStatus", "machine", klog.KObj(machineObj))

	tick := timeClock.NewTicker(machineStatusCheckInterval)
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
