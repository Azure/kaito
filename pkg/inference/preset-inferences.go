// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package inference

import (
	"context"
	"fmt"
	"strconv"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/resources"
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
			Key:      resources.GPUString,
		},
		{
			Effect: corev1.TaintEffectNoSchedule,
			Value:  resources.GPUString,
			Key:    "sku",
		},
	}
)

func updateTorchParamsForDistributedInference(ctx context.Context, kubeClient client.Client, wObj *kaitov1alpha1.Workspace, inferenceObj *model.PresetInferenceParam) error {
	existingService := &corev1.Service{}
	err := resources.GetResource(ctx, wObj.Name, wObj.Namespace, kubeClient, existingService)
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
	if inferenceObj.TorchRunRdzvParams != nil {
		inferenceObj.TorchRunRdzvParams["max_restarts"] = "3"
		inferenceObj.TorchRunRdzvParams["rdzv_id"] = "job"
		inferenceObj.TorchRunRdzvParams["rdzv_backend"] = "c10d"
		inferenceObj.TorchRunRdzvParams["rdzv_endpoint"] =
			fmt.Sprintf("%s-0.%s-headless.%s.svc.cluster.local:29500", wObj.Name, wObj.Name, wObj.Namespace)
	}
	return nil
}

func CreatePresetInference(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	inferenceObj *model.PresetInferenceParam, supportDistributedInference bool, kubeClient client.Client) (client.Object, error) {
	if inferenceObj.TorchRunParams != nil && supportDistributedInference {
		if err := updateTorchParamsForDistributedInference(ctx, kubeClient, workspaceObj, inferenceObj); err != nil {
			klog.ErrorS(err, "failed to update torch params", "workspace", workspaceObj)
			return nil, err
		}
	}

	volume, volumeMount := configVolume(workspaceObj, inferenceObj)
	commands, resourceReq := prepareInferenceParameters(ctx, inferenceObj)

	var depObj client.Object
	if supportDistributedInference {
		depObj = resources.GenerateStatefulSetManifest(ctx, workspaceObj, inferenceObj.Image, inferenceObj.ImagePullSecrets, *workspaceObj.Resource.Count, commands,
			containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volume, volumeMount)
	} else {
		depObj = resources.GenerateDeploymentManifest(ctx, workspaceObj, inferenceObj.Image, inferenceObj.ImagePullSecrets, *workspaceObj.Resource.Count, commands,
			containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volume, volumeMount)
	}
	err := resources.CreateResource(ctx, depObj, kubeClient)
	if client.IgnoreAlreadyExists(err) != nil {
		return nil, err
	}
	return depObj, nil
}

// prepareInferenceParameters builds a PyTorch command:
// torchrun <TORCH_PARAMS> <OPTIONAL_RDZV_PARAMS> baseCommand <MODEL_PARAMS>
// and sets the GPU resources required for inference.
// Returns the command and resource configuration.
func prepareInferenceParameters(ctx context.Context, inferenceObj *model.PresetInferenceParam) ([]string, corev1.ResourceRequirements) {
	torchCommand := buildCommandStr(inferenceObj.BaseCommand, inferenceObj.TorchRunParams)
	torchCommand = buildCommandStr(torchCommand, inferenceObj.TorchRunRdzvParams)
	modelCommand := buildCommandStr(inferenceObj.InferenceFile, inferenceObj.ModelRunParams)
	commands := shellCommand(torchCommand + " " + modelCommand)

	resourceRequirements := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceName(resources.CapacityNvidiaGPU): resource.MustParse(inferenceObj.GPUCountRequirement),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceName(resources.CapacityNvidiaGPU): resource.MustParse(inferenceObj.GPUCountRequirement),
		},
	}

	return commands, resourceRequirements
}

func configVolume(wObj *kaitov1alpha1.Workspace, inferenceObj *model.PresetInferenceParam) ([]corev1.Volume, []corev1.VolumeMount) {
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
