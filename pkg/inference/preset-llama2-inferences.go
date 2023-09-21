package inference

import (
	"context"
	"fmt"
	"time"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/kdm/pkg/k8sresources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Preset2ATimeout defines the maximum duration for pulling the PresetA image.
	// This timeout accommodates the size of PresetA, ensuring pull completion
	// even under slower network conditions or unforeseen delays.
	Preset2ATimeout = time.Duration(10) * time.Minute

	// Preset2BTimeout defines the maximum duration for pulling the PresetB image.
	// This timeout accommodates the size of PresetB.
	Preset2BTimeout = time.Duration(20) * time.Minute

	// Preset2CTimeout defines the maximum duration for pulling the PresetC image.
	// This timeout accommodates the size of PresetC (the largest image).
	Preset2CTimeout = time.Duration(30) * time.Minute

	RegistryName                   = "aimodelsregistry.azurecr.io"
	PresetSetModelllama2AChatImage = RegistryName + "/llama-2-7b-chat:latest"
	PresetSetModelllama2BChatImage = RegistryName + "/llama-2-13b-chat:latest"
	PresetSetModelllama2CChatImage = RegistryName + "/llama-2-70b-chat:latest"

	ProbePath = "/healthz"
	Port5000  = int32(5000)

	BaseCommandPresetSetModelllama2A = "cd /workspace/llama/llama-2-7b-chat && torchrun"
	BaseCommandPresetSetModelllama2B = "cd /workspace/llama/llama-2-13b-chat && torchrun"
	BaseCommandPresetSetModelllama2C = "cd /workspace/llama/llama-2-70b-chat && torchrun"
	PythonModelInferenceServerFile   = "web_example_chat_completion.py"
)

var llamaRunParams = map[string]string{
	"max_seq_len":    "512",
	"max_batch_size": "8",
}

var (
	containerPorts = []corev1.ContainerPort{{
		ContainerPort: Port5000,
	},
	}

	livenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromInt(5000),
				Path: ProbePath,
			},
		},
		InitialDelaySeconds: 600, // 10 minutes
		PeriodSeconds:       10,
	}

	readinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"./llama-readiness-check.sh"},
			},
		},
		InitialDelaySeconds: 30,
		PeriodSeconds:       10,
	}

	tolerations = []corev1.Toleration{
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Operator: corev1.TolerationOpEqual,
			Key:      k8sresources.GPUString,
		},
		{
			Effect: corev1.TaintEffectNoSchedule,
			Value:  k8sresources.GPUString,
			Key:    "sku",
		},
	}
)

func CreateLLAMA2APresetModel(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace,
	torchRunParams map[string]string, kubeClient client.Client) error {
	klog.InfoS("CreateLLAMA2APresetModel", "workspace", klog.KObj(workspaceObj))
	commands := buildCommandStr(BaseCommandPresetSetModelllama2A, torchRunParams)
	commands += " " + PythonModelInferenceServerFile
	shellCmd := shellCommand(buildCommandStr(commands, llamaRunParams))
	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse("1"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse("1"),
		},
	}

	// Replica is always 1, because LLAMA2APreset only runs on one GPU
	depObj := k8sresources.GenerateStatefulSetManifest(ctx, workspaceObj, PresetSetModelllama2AChatImage,
		1, shellCmd, containerPorts, livenessProbe, readinessProbe, resourceRequirements, tolerations)
	err := k8sresources.CreateResource(ctx, depObj, kubeClient)
	if err != nil {
		return err
	}

	if err := checkResourceStatus(depObj, kubeClient, Preset2ATimeout); err != nil {
		return err
	}
	return nil
}

func CreateLLAMA2BPresetModel(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace,
	torchRunParams map[string]string, kubeClient client.Client) error {
	klog.InfoS("CreateLLAMA2BPresetModel", "workspace", klog.KObj(workspaceObj))

	commands := buildCommandStr(BaseCommandPresetSetModelllama2B, torchRunParams)
	commands += " " + PythonModelInferenceServerFile
	shellCmd := shellCommand(buildCommandStr(commands, llamaRunParams))

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse(torchRunParams["nproc_per_node"]),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse(torchRunParams["nproc_per_node"]),
		},
	}

	depObj := k8sresources.GenerateStatefulSetManifest(ctx, workspaceObj, PresetSetModelllama2BChatImage,
		*workspaceObj.Resource.Count, shellCmd, containerPorts, livenessProbe, readinessProbe, resourceRequirements, tolerations)

	if err := k8sresources.CreateResource(ctx, depObj, kubeClient); err != nil {
		return err
	}

	if err := checkResourceStatus(depObj, kubeClient, Preset2BTimeout); err != nil {
		return err
	}
	return nil
}

func CreateLLAMA2CPresetModel(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace,
	torchRunParams map[string]string, kubeClient client.Client) error {
	klog.InfoS("CreateLLAMA2CPresetModel", "workspace", klog.KObj(workspaceObj))

	commands := buildCommandStr(BaseCommandPresetSetModelllama2C, torchRunParams)
	commands += " " + PythonModelInferenceServerFile
	shellCmd := shellCommand(buildCommandStr(commands, llamaRunParams))

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse(torchRunParams["nproc_per_node"]),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse(torchRunParams["nproc_per_node"]),
			corev1.ResourceEphemeralStorage:                     resource.MustParse("300Gi"),
		},
	}

	depObj := k8sresources.GenerateStatefulSetManifest(ctx, workspaceObj, PresetSetModelllama2CChatImage,
		*workspaceObj.Resource.Count, shellCmd, containerPorts, livenessProbe, readinessProbe, resourceRequirements, tolerations)

	if err := k8sresources.CreateResource(ctx, depObj, kubeClient); err != nil {
		return err
	}

	if err := checkResourceStatus(depObj, kubeClient, Preset2CTimeout); err != nil {
		return err
	}
	return nil
}

func checkResourceStatus(obj client.Object, kubeClient client.Client, timeoutDuration time.Duration) error {
	klog.InfoS("checkResourceStatus", "resource", obj.GetName())

	// Use Context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			key := client.ObjectKey{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			}
			err := kubeClient.Get(ctx, key, obj)
			if err != nil {
				return err
			}

			switch resource := obj.(type) {
			case *appsv1.Deployment:
				if resource.Status.ReadyReplicas == *resource.Spec.Replicas {
					klog.InfoS("deployment status is ready", "deployment", resource.Name)
					return nil
				}
			case *appsv1.StatefulSet:
				if resource.Status.ReadyReplicas == *resource.Spec.Replicas {
					klog.InfoS("statefulset status is ready", "statefulset", resource.Name)
					return nil
				}
			default:
				return fmt.Errorf("unsupported resource type")
			}
		}
	}
}

func shellCommand(command string) []string {
	return []string{
		"/bin/sh",
		"-c",
		command,
	}
}

func buildCommandStr(baseCommand string, torchRunParams map[string]string) string {
	updatedBaseCommand := baseCommand
	for key, value := range torchRunParams {
		updatedBaseCommand = fmt.Sprintf("%s --%s=%s", updatedBaseCommand, key, value)
	}

	return updatedBaseCommand
}
