// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package inference

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/kaito-project/kaito/pkg/utils"
	"github.com/kaito-project/kaito/pkg/utils/consts"

	"github.com/kaito-project/kaito/api/v1alpha1"
	kaitov1alpha1 "github.com/kaito-project/kaito/api/v1alpha1"
	"github.com/kaito-project/kaito/pkg/model"
	"github.com/kaito-project/kaito/pkg/utils/resources"
	"github.com/kaito-project/kaito/pkg/workspace/manifests"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProbePath = "/health"
	Port5000  = 5000
)

var (
	containerPorts = []corev1.ContainerPort{{
		ContainerPort: int32(Port5000),
	},
	}

	livenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromInt(Port5000),
				Path: ProbePath,
			},
		},
		InitialDelaySeconds: 600, // 10 minutes
		PeriodSeconds:       10,
	}

	readinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromInt(Port5000),
				Path: ProbePath,
			},
		},
		InitialDelaySeconds: 30,
		PeriodSeconds:       10,
	}

	tolerations = []corev1.Toleration{
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Operator: corev1.TolerationOpExists,
			Key:      resources.CapacityNvidiaGPU,
		},
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Value:    consts.GPUString,
			Key:      consts.SKUString,
			Operator: corev1.TolerationOpEqual,
		},
	}
)

func updateTorchParamsForDistributedInference(ctx context.Context, kubeClient client.Client, wObj *kaitov1alpha1.Workspace, inferenceParam *model.PresetParam) error {
	runtimeName := v1alpha1.GetWorkspaceRuntimeName(wObj)
	if runtimeName != model.RuntimeNameHuggingfaceTransformers {
		return fmt.Errorf("distributed inference is not supported for runtime %s", runtimeName)
	}

	existingService := &corev1.Service{}
	err := resources.GetResource(ctx, wObj.Name, wObj.Namespace, kubeClient, existingService)
	if err != nil {
		return err
	}

	nodes := *wObj.Resource.Count
	inferenceParam.Transformers.TorchRunParams["nnodes"] = strconv.Itoa(nodes)
	inferenceParam.Transformers.TorchRunParams["nproc_per_node"] = strconv.Itoa(inferenceParam.WorldSize / nodes)
	if nodes > 1 {
		inferenceParam.Transformers.TorchRunParams["node_rank"] = "$(echo $HOSTNAME | grep -o '[^-]*$')"
		inferenceParam.Transformers.TorchRunParams["master_addr"] = existingService.Spec.ClusterIP
		inferenceParam.Transformers.TorchRunParams["master_port"] = "29500"
	}
	if inferenceParam.Transformers.TorchRunRdzvParams != nil {
		inferenceParam.Transformers.TorchRunRdzvParams["max_restarts"] = "3"
		inferenceParam.Transformers.TorchRunRdzvParams["rdzv_id"] = "job"
		inferenceParam.Transformers.TorchRunRdzvParams["rdzv_backend"] = "c10d"
		inferenceParam.Transformers.TorchRunRdzvParams["rdzv_endpoint"] =
			fmt.Sprintf("%s-0.%s-headless.%s.svc.cluster.local:29500", wObj.Name, wObj.Name, wObj.Namespace)
	}
	return nil
}

func GetInferenceImageInfo(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, presetObj *model.PresetParam) (string, []corev1.LocalObjectReference) {
	imagePullSecretRefs := []corev1.LocalObjectReference{}
	// Check if the workspace preset's access mode is private
	if len(workspaceObj.Inference.Adapters) > 0 {
		for _, adapter := range workspaceObj.Inference.Adapters {
			for _, secretName := range adapter.Source.ImagePullSecrets {
				imagePullSecretRefs = append(imagePullSecretRefs, corev1.LocalObjectReference{Name: secretName})
			}
		}
	}
	if string(workspaceObj.Inference.Preset.AccessMode) == string(kaitov1alpha1.ModelImageAccessModePrivate) {
		imageName := workspaceObj.Inference.Preset.PresetOptions.Image
		for _, secretName := range workspaceObj.Inference.Preset.PresetOptions.ImagePullSecrets {
			imagePullSecretRefs = append(imagePullSecretRefs, corev1.LocalObjectReference{Name: secretName})
		}
		return imageName, imagePullSecretRefs
	} else {
		imageName := string(workspaceObj.Inference.Preset.Name)
		imageTag := presetObj.Tag
		registryName := os.Getenv("PRESET_REGISTRY_NAME")
		imageName = fmt.Sprintf("%s/kaito-%s:%s", registryName, imageName, imageTag)

		return imageName, imagePullSecretRefs
	}
}

func CreatePresetInference(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, revisionNum string,
	model model.Model, kubeClient client.Client) (client.Object, error) {
	inferenceParam := model.GetInferenceParameters().DeepCopy()

	configVolume, err := EnsureInferenceConfigMap(ctx, workspaceObj, kubeClient)
	if err != nil {
		return nil, err
	}

	if model.SupportDistributedInference() {
		if err := updateTorchParamsForDistributedInference(ctx, kubeClient, workspaceObj, inferenceParam); err != nil {
			klog.ErrorS(err, "failed to update torch params", "workspace", workspaceObj)
			return nil, err
		}
	}

	// resource requirements
	skuNumGPUs, err := utils.GetSKUNumGPUs(ctx, kubeClient, workspaceObj.Status.WorkerNodes,
		workspaceObj.Resource.InstanceType, inferenceParam.GPUCountRequirement)
	if err != nil {
		return nil, fmt.Errorf("failed to get SKU num GPUs: %v", err)
	}
	resourceReq := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceName(resources.CapacityNvidiaGPU): resource.MustParse(skuNumGPUs),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceName(resources.CapacityNvidiaGPU): resource.MustParse(skuNumGPUs),
		},
	}
	skuGPUCount, _ := strconv.Atoi(skuNumGPUs)

	// additional volume
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	// Add config volume mount
	cmVolume, cmVolumeMount := utils.ConfigCMVolume(configVolume.Name)
	volumes = append(volumes, cmVolume)
	volumeMounts = append(volumeMounts, cmVolumeMount)

	// add share memory for cross process communication
	shmVolume, shmVolumeMount := utils.ConfigSHMVolume(skuGPUCount)
	if shmVolume.Name != "" {
		volumes = append(volumes, shmVolume)
	}
	if shmVolumeMount.Name != "" {
		volumeMounts = append(volumeMounts, shmVolumeMount)
	}
	if len(workspaceObj.Inference.Adapters) > 0 {
		adapterVolume, adapterVolumeMount := utils.ConfigAdapterVolume()
		volumes = append(volumes, adapterVolume)
		volumeMounts = append(volumeMounts, adapterVolumeMount)
	}

	// inference command
	runtimeName := kaitov1alpha1.GetWorkspaceRuntimeName(workspaceObj)
	commands := inferenceParam.GetInferenceCommand(runtimeName, skuNumGPUs, &cmVolumeMount)

	image, imagePullSecrets := GetInferenceImageInfo(ctx, workspaceObj, inferenceParam)

	var depObj client.Object
	if model.SupportDistributedInference() {
		depObj = manifests.GenerateStatefulSetManifest(ctx, workspaceObj, image, imagePullSecrets, *workspaceObj.Resource.Count, commands,
			containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volumes, volumeMounts)
	} else {
		depObj = manifests.GenerateDeploymentManifest(ctx, workspaceObj, revisionNum, image, imagePullSecrets, *workspaceObj.Resource.Count, commands,
			containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volumes, volumeMounts)
	}
	err = resources.CreateResource(ctx, depObj, kubeClient)
	if client.IgnoreAlreadyExists(err) != nil {
		return nil, err
	}
	return depObj, nil
}

// EnsureInferenceConfigMap handles two scenarios:
// 1. User provided config (workspaceObj.Inference.Config):
//   - Check if it exists in the target namespace
//   - If not found, return error as this is user-specified
//
// 2. No user config specified:
//   - Use the default config template (inference-params-template)
//   - Check if it exists in the target namespace
//   - If not, copy from release namespace to target namespace
func EnsureInferenceConfigMap(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	kubeClient client.Client) (*corev1.ConfigMap, error) {

	// If user specified a config, use that
	if workspaceObj.Inference.Config != "" {
		userCM := &corev1.ConfigMap{}
		err := resources.GetResource(ctx, workspaceObj.Inference.Config, workspaceObj.Namespace, kubeClient, userCM)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("user specified ConfigMap %s not found in namespace %s",
					workspaceObj.Inference.Config, workspaceObj.Namespace)
			}
			return nil, err
		}

		return userCM, nil
	}

	// Otherwise use default template
	configMapName := kaitov1alpha1.DefaultInferenceConfigTemplate

	// Check if default configmap already exists in target namespace
	existingCM := &corev1.ConfigMap{}
	err := resources.GetResource(ctx, configMapName, workspaceObj.Namespace, kubeClient, existingCM)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
	} else {
		klog.Infof("Default ConfigMap already exists in target namespace: %s, no action taken.", workspaceObj.Namespace)
		return existingCM, nil
	}

	// Copy default template from release namespace if not found
	releaseNamespace, err := utils.GetReleaseNamespace()
	if err != nil {
		return nil, fmt.Errorf("failed to get release namespace: %v", err)
	}

	templateCM := &corev1.ConfigMap{}
	err = resources.GetResource(ctx, configMapName, releaseNamespace, kubeClient, templateCM)
	if err != nil {
		return nil, fmt.Errorf("failed to get default ConfigMap from template namespace: %v", err)
	}

	templateCM.Namespace = workspaceObj.Namespace
	templateCM.ResourceVersion = "" // Clear metadata not needed for creation
	templateCM.UID = ""             // Clear UID

	err = resources.CreateResource(ctx, templateCM, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create default ConfigMap in target namespace %s: %v",
			workspaceObj.Namespace, err)
	}

	return templateCM, nil
}
