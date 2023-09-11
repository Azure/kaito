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
	err := k8sresources.CreateDeployment(ctx, depObj, kubeClient)
	if err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := checkResourceStatus(ctxWithTimeout, depObj, kubeClient); err != nil {
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
		1, commands, containerPorts, livenessProbe, readinessProbe, resourceRequirements, volumeMount, tolerations, volume)

	if err := k8sresources.CreateDeployment(ctx, depObj, kubeClient); err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	if err := checkResourceStatus(ctxWithTimeout, depObj, kubeClient); err != nil {
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

	depObj := k8sresources.GenerateDeploymentManifest(ctx, workspaceObj, PresetSetModelllama2CChatImage,
		1, commands, containerPorts, livenessProbe, readinessProbe, resourceRequirements, volumeMount, tolerations, volume)

	if err := k8sresources.CreateDeployment(ctx, depObj, kubeClient); err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := checkResourceStatus(ctxWithTimeout, depObj, kubeClient); err != nil {
		return err
	}
	return nil
}

func checkResourceStatus(ctx context.Context, obj client.Object, kubeClient client.Client) error {
	klog.InfoS("checkResourceStatus", "resource", obj.GetName())
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Use context for timeout
	timeoutChan := ctx.Done()

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("check resource status timed out. resource %s is not ready", obj.GetName())

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
