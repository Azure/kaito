// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package inference

import (
	"context"
	"fmt"
	"strconv"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
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

func setTorchParams(ctx context.Context, kubeClient client.Client, wObj *kaitov1alpha1.Workspace, inferenceObj PresetInferenceParam) error {
	if inferenceObj.ModelName == "LLaMa2" {
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
				fmt.Sprintf("%s-0.%s-headless.default.svc.cluster.local:29500", wObj.Name, wObj.Name)
		}
	} else if inferenceObj.ModelName == "Falcon" {
		inferenceObj.TorchRunParams["config_file"] = "config.yaml"
		inferenceObj.TorchRunParams["num_processes"] = "1"
		inferenceObj.TorchRunParams["num_machines"] = "1"
		inferenceObj.TorchRunParams["machine_rank"] = "0"
		inferenceObj.TorchRunParams["gpu_ids"] = "all"
	}
	return nil
}

func CreatePresetInference(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	inferenceObj PresetInferenceParam, useHeadlessService bool, kubeClient client.Client) (client.Object, error) {
	if inferenceObj.TorchRunParams != nil {
		if err := setTorchParams(ctx, kubeClient, workspaceObj, inferenceObj); err != nil {
			klog.ErrorS(err, "failed to update torch params", "workspace", workspaceObj)
			return nil, err
		}
	}

	volume, volumeMount := configVolume(workspaceObj, inferenceObj)
	commands, resourceReq := prepareInferenceParameters(ctx, inferenceObj)

	var depObj client.Object
	switch inferenceObj.ModelName {
	case "LLaMa2":
		depObj = resources.GenerateStatefulSetManifest(ctx, workspaceObj, inferenceObj.Image, inferenceObj.ImagePullSecrets, *workspaceObj.Resource.Count, commands,
			containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volume, volumeMount, useHeadlessService)
	case "Falcon":
		depObj = resources.GenerateDeploymentManifest(ctx, workspaceObj, inferenceObj.Image, inferenceObj.ImagePullSecrets, *workspaceObj.Resource.Count, commands,
			containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volume, volumeMount)
	default:
		return nil, fmt.Errorf("Model not recognized: %s", inferenceObj.ModelName)
	}
	err := resources.CreateResource(ctx, depObj, kubeClient)
	if client.IgnoreAlreadyExists(err) != nil {
		return nil, err
	}
	return depObj, nil
}

func prepareInferenceParameters(ctx context.Context, inferenceObj PresetInferenceParam) ([]string, corev1.ResourceRequirements) {
	torchCommand := buildCommandStr(inferenceObj.BaseCommand, inferenceObj.TorchRunParams)
	torchCommand = buildCommandStr(torchCommand, inferenceObj.TorchRunRdzvParams)
	modelCommand := buildCommandStr(inferenceObj.InferenceFile, inferenceObj.ModelRunParams)
	commands := shellCommand(torchCommand + " " + modelCommand)

	resourceRequirements := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceName(resources.CapacityNvidiaGPU): resource.MustParse(inferenceObj.GPURequirement),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceName(resources.CapacityNvidiaGPU): resource.MustParse(inferenceObj.GPURequirement),
		},
	}

	return commands, resourceRequirements
}

func configVolume(wObj *kaitov1alpha1.Workspace, inferenceObj PresetInferenceParam) ([]corev1.Volume, []corev1.VolumeMount) {
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
