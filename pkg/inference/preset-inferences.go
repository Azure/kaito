package inference

import (
	"context"
	"fmt"
	"strconv"
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

func setTorchParams(ctx context.Context, kubeClient client.Client, wObj *kdmv1alpha1.Workspace, inferenceObj PresetInferenceParam) error {
	if inferenceObj.ModelName == "LLaMa2" {
		existingService := &corev1.Service{}
		err := k8sresources.GetResource(ctx, wObj.Name, wObj.Namespace, kubeClient, existingService)
		if err != nil {
			return err
		}

		nodes := *wObj.Resource.Count
		inferenceObj.TorchRunParams["nnodes"] = strconv.Itoa(nodes)
		inferenceObj.TorchRunParams["nproc_per_node"] = strconv.Itoa(inferenceObj.WorldSize / nodes)
		if nodes > 1 {
			inferenceObj.TorchRunParams["node_rank"] = "$(echo $HOSTNAME | grep -o '[^-]*$')"
			inferenceObj.TorchRunParams["master_addr"] = existingService.Spec.ClusterIP
			inferenceObj.TorchRunParams["master_port"] = "29500"
		}
	}
	return nil
}

func CreatePresetInference(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace,
	inferenceObj PresetInferenceParam, kubeClient client.Client) error {
	klog.InfoS("CreatePresetInference", "workspace", klog.KObj(workspaceObj))

	if inferenceObj.TorchRunParams != nil {
		if err := setTorchParams(ctx, kubeClient, workspaceObj, inferenceObj); err != nil {
			klog.ErrorS(err, "failed to update torch params", "workspace", workspaceObj)
			return err
		}
	}

	volume, volumeMount := configVolume(workspaceObj, inferenceObj)
	commands, resourceReq := prepareInferenceParameters(ctx, inferenceObj)

	depObj := k8sresources.GenerateStatefulSetManifest(ctx, workspaceObj, inferenceObj.Image, *workspaceObj.Resource.Count, commands,
		containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volume, volumeMount)
	err := k8sresources.CreateResource(ctx, depObj, kubeClient)
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

func prepareInferenceParameters(ctx context.Context, inferenceObj PresetInferenceParam) ([]string, corev1.ResourceRequirements) {
	torchCommand := buildCommandStr(inferenceObj.BaseCommand, inferenceObj.TorchRunParams)
	modelCommand := buildCommandStr(inferenceObj.InferenceFile, inferenceObj.ModelRunParams)
	commands := shellCommand(torchCommand + " " + modelCommand)

	resourceRequirements := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse(inferenceObj.GPURequirement),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceName(k8sresources.CapacityNvidiaGPU): resource.MustParse(inferenceObj.GPURequirement),
		},
	}

	return commands, resourceRequirements
}

func configVolume(wObj *kdmv1alpha1.Workspace, inferenceObj PresetInferenceParam) ([]corev1.Volume, []corev1.VolumeMount) {
	volume := []corev1.Volume{}
	volumeMount := []corev1.VolumeMount{}

	// Signifies multinode inference requirement
	if *wObj.Resource.Count > 1 {
		// Append share memory volume to any existing volumes
		volume = append(volume, corev1.Volume{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: "Memory",
				},
			},
		})

		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      volume[0].Name,
			MountPath: inferenceObj.DefaultVolumeMountPath,
		})
	}

	return volume, volumeMount
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
