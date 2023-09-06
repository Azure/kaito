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
	"k8s.io/utils/clock"
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
	deploymentStatusCheckInterval = 600 * time.Second

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
	if err := checkDeploymentStatus(ctx, depObj, kubeClient); err != nil {
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

	if err := checkDeploymentStatus(ctx, depObj, kubeClient); err != nil {
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

	if err := checkDeploymentStatus(ctx, depObj, kubeClient); err != nil {
		return err
	}
	return nil
}

func checkDeploymentStatus(ctx context.Context, depObj *appsv1.Deployment, kubeClient client.Client) error {
	klog.InfoS("checkDeploymentStatus", "deployment", depObj.Name)
	timeClock := clock.RealClock{}
	tick := timeClock.NewTicker(deploymentStatusCheckInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-tick.C():
			return fmt.Errorf("check deployment status timed out. deployment %s is not ready", depObj.Name)
		default:
			time.Sleep(1 * time.Second)
			err := kubeClient.Get(ctx, client.ObjectKey{
				Name:      depObj.Name,
				Namespace: depObj.Namespace,
			}, depObj)
			if err != nil {
				return err
			}
			if depObj.Status.ReadyReplicas != 1 {
				continue
			}

			klog.InfoS("inference deployment status is ready", "deployment", depObj.Name)
			return nil
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
