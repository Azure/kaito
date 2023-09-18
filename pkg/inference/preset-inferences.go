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
	ProbePath = "/healthz"
	Port5000  = int32(5000)
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

func CreatePresetInference(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace, volume []corev1.Volume,
	inferenceObj PresetInferenceParam, kubeClient client.Client) error {
	klog.InfoS("CreatePresetInference", "workspace", klog.KObj(workspaceObj))

	commands, resourceReq, volumeMount := prepareInferenceParameters(torchRunParams, volume, inferenceObj)

	depObj := k8sresources.GenerateDeploymentManifest(ctx, workspaceObj, inferenceObj.Image, 1, commands,
		containerPorts, livenessProbe, readinessProbe, resourceReq, volumeMount, tolerations, volume)
	err := k8sresources.CreateDeployment(ctx, depObj, kubeClient)
	if err != nil {
		return err
	}

	if err := checkResourceStatus(depObj, kubeClient, inferenceObj.DeploymentTimeout); err != nil {
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

			switch k8sResource := obj.(type) {
			case *appsv1.Deployment:
				if k8sResource.Status.ReadyReplicas == *k8sResource.Spec.Replicas {
					klog.InfoS("deployment status is ready", "deployment", k8sResource.Name)
					return nil
				}
			case *appsv1.StatefulSet:
				if k8sResource.Status.ReadyReplicas == *k8sResource.Spec.Replicas {
					klog.InfoS("statefulset status is ready", "statefulset", k8sResource.Name)
					return nil
				}
			default:
				return fmt.Errorf("unsupported resource type")
			}
		}
	}
}

func prepareInferenceParameters(torchRunParams map[string]string, volume []corev1.Volume,
	inferenceObj PresetInferenceParam) ([]string, corev1.ResourceRequirements, []corev1.VolumeMount) {
	commands := buildCommand(inferenceObj.BaseCommand, torchRunParams)

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse(inferenceObj.GPURequirement),
			corev1.ResourceStorage:                              resource.MustParse(inferenceObj.DiskStorageRequirement),
			corev1.ResourceRequestsMemory:                       resource.MustParse(inferenceObj.GPUMemoryRequirement),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse(inferenceObj.GPURequirement),
			corev1.ResourceStorage:                              resource.MustParse(inferenceObj.DiskStorageRequirement),
			corev1.ResourceRequestsMemory:                       resource.MustParse(inferenceObj.GPUMemoryRequirement),
		},
	}
	volumeMount := []corev1.VolumeMount{}
	if len(volume) != 0 {
		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      volume[0].Name,
			MountPath: inferenceObj.DefaultVolumeMountPath,
		})
	}
	return commands, resourceRequirements, volumeMount
}

func buildCommand(baseCommand string, torchRunParams map[string]string) []string {
	updatedBaseCommand := baseCommand
	for key, value := range torchRunParams {
		updatedBaseCommand = fmt.Sprintf("%s --%s=%s", updatedBaseCommand, key, value)
	}

	commands := []string{
		"/bin/sh",
		"-c",
		updatedBaseCommand,
	}

	return commands
}
