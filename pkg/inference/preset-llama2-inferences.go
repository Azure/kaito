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
	// Preset2ATimeout Preset2BTimeout Preset2CTimeout define maximum durations for pulling image presets.
	// Durations increase from PresetA (smallest image) to PresetC (largest). These timeouts accommodate large image sizes,
	// ensuring pull completion even under slower network conditions or unforeseen delays.
	Preset2ATimeout = time.Duration(10) * time.Minute
	Preset2BTimeout = time.Duration(20) * time.Minute
	Preset2CTimeout = time.Duration(30) * time.Minute

	RegistryName                   = "aimodelsregistry.azurecr.io"
	PresetSetModelllama2AChatImage = RegistryName + "/llama-2-7b-chat:latest"
	PresetSetModelllama2BChatImage = RegistryName + "/llama-2-13b-chat:latest"
	PresetSetModelllama2CChatImage = RegistryName + "/llama-2-70b-chat:latest"

	ProbePath = "/healthz"
	Port5000  = int32(5000)

	BaseCommandPresetSetModelllama2A = "cd /workspace/llama/llama-2-7b-chat && torchrun web_example_chat_completion.py"
	BaseCommandPresetSetModelllama2B = "cd /workspace/llama/llama-2-13b-chat && torchrun --nproc_per_node=2 web_example_chat_completion.py"
	BaseCommandPresetSetModelllama2C = "cd /workspace/llama/llama-2-70b-chat && torchrun --nproc_per_node=4 web_example_chat_completion.py"
)

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
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromInt(5000),
				Path: ProbePath,
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

func CreateLLAMA2APresetModel(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace, volume []corev1.Volume,
	torchRunParams map[string]string, kubeClient client.Client) error {
	klog.InfoS("CreateLLAMA2APresetModel", "workspace", klog.KObj(workspaceObj))
	commands := buildCommand(BaseCommandPresetSetModelllama2A, torchRunParams)
	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse("1"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse("1"),
		},
	}
	volumeMount := []corev1.VolumeMount{}
	if len(volume) != 0 {
		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      volume[0].Name,
			MountPath: "/dev/shm",
		})
	}

	depObj := k8sresources.GenerateDeploymentManifest(ctx, workspaceObj, PresetSetModelllama2AChatImage,
		1, commands, containerPorts, livenessProbe, readinessProbe, resourceRequirements, volumeMount, tolerations, volume)
	err := k8sresources.CreateResource(ctx, depObj, kubeClient)
	if err != nil {
		return err
	}

	if err := checkResourceStatus(depObj, kubeClient, Preset2ATimeout); err != nil {
		return err
	}
	return nil
}

func CreateLLAMA2BPresetModel(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace, volume []corev1.Volume,
	torchRunParams map[string]string, kubeClient client.Client) error {
	klog.InfoS("CreateLLAMA2BPresetModel", "workspace", klog.KObj(workspaceObj))

	commands := buildCommand(BaseCommandPresetSetModelllama2B, torchRunParams)

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse("2"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse("2"),
		},
	}
	volumeMount := []corev1.VolumeMount{}
	if len(volume) != 0 {
		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      volume[0].Name,
			MountPath: "/dev/shm",
		})
	}

	depObj := k8sresources.GenerateDeploymentManifest(ctx, workspaceObj, PresetSetModelllama2BChatImage,
		1 /*TODO: PARAM TO BE SET BY USER*/, commands, containerPorts, livenessProbe, readinessProbe, resourceRequirements, volumeMount, tolerations, volume)

	if err := k8sresources.CreateResource(ctx, depObj, kubeClient); err != nil {
		return err
	}

	if err := checkResourceStatus(depObj, kubeClient, Preset2BTimeout); err != nil {
		return err
	}
	return nil
}

func CreateLLAMA2CPresetModel(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace, volume []corev1.Volume,
	torchRunParams map[string]string, kubeClient client.Client) error {
	klog.InfoS("CreateLLAMA2CPresetModel", "workspace", klog.KObj(workspaceObj))
	commands := buildCommand(BaseCommandPresetSetModelllama2C, torchRunParams)

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse("4"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse("4"),
			corev1.ResourceEphemeralStorage:                     resource.MustParse("300Gi"),
		},
	}

	volumeMount := []corev1.VolumeMount{}
	if len(volume) != 0 {
		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      volume[0].Name,
			MountPath: "/dev/shm",
		})
	}

	depObj := k8sresources.GenerateStatefulSetManifest(ctx, workspaceObj, PresetSetModelllama2CChatImage,
		1, commands, containerPorts, livenessProbe, readinessProbe, resourceRequirements, volumeMount, tolerations, volume)

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

func buildCommand(baseCommand string, torchRunParams map[string]string) []string {
	var updatedBaseCommand string
	for key, value := range torchRunParams {
		updatedBaseCommand = fmt.Sprintf("%s --%s=%s", baseCommand, key, value)
	}

	commands := []string{
		"/bin/sh",
		"-c",
		updatedBaseCommand,
	}

	return commands
}
