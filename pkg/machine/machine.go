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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProvisionerName           = "default"
	LabelGPUProvisionerCustom = "gpu-provisioner.sh/machine-type"
	LabelProvisionerName      = "karpenter.sh/provisioner-name"
	GPUString                 = "gpu"
)

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

func CreateMachine(ctx context.Context, machineObj *v1alpha5.Machine, kubeClient client.Client) error {
	klog.InfoS("CreateMachine", "machine", klog.KObj(machineObj))
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Create(ctx, machineObj, &client.CreateOptions{})
	})
}

func CheckMachineStatus(ctx context.Context, machineObj *v1alpha5.Machine, kubeClient client.Client) error {
	klog.InfoS("CheckMachineStatus", "machine", klog.KObj(machineObj))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(1 * time.Second)
			err := kubeClient.Get(ctx, client.ObjectKey{Name: machineObj.Name, Namespace: machineObj.Namespace}, machineObj, &client.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				} else {
					return err
				}
			}
			_, found := lo.Find(machineObj.GetConditions(), func(condition apis.Condition) bool {
				return condition.Type == apis.ConditionReady &&
					condition.Status == v1.ConditionTrue
			})
			if !found {
				continue
			}
			klog.InfoS("machine status is ready", "machine", machineObj.Name)
			return nil
		}
	}
}
