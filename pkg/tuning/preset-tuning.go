package tuning

import (
	"context"
	"fmt"
	"os"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/resources"
	"github.com/azure/kaito/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProbePath  = "/healthz"
	Port5000   = int32(5000)
	TuningFile = "fine_tuning_api.py"
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

func getInstanceGPUCount(sku string) int {
	gpuConfig, exists := kaitov1alpha1.SupportedGPUConfigs[sku]
	if !exists {
		return 1
	}
	return gpuConfig.GPUCount
}

func formatRegistryImagePath(registryName, imageName, imageTag string) string {
	if imageTag != "" {
		return fmt.Sprintf("%s/%s:%s", registryName, imageName, imageTag)
	}
	return fmt.Sprintf("%s/%s", registryName, imageName)
}

func GetTuningImageInfo(ctx context.Context, wObj *kaitov1alpha1.Workspace, presetObj *model.PresetParam) string {
	registryName := os.Getenv("PRESET_REGISTRY_NAME")
	return formatRegistryImagePath(registryName, "kaito-tuning-"+string(wObj.Tuning.Preset.Name), presetObj.Tag)
}

func GetDataSrcImageInfo(ctx context.Context, wObj *kaitov1alpha1.Workspace) (string, []corev1.LocalObjectReference) {
	imagePullSecretRefs := make([]corev1.LocalObjectReference, len(wObj.Tuning.Input.ImagePullSecrets))
	for i, secretName := range wObj.Tuning.Input.ImagePullSecrets {
		imagePullSecretRefs[i] = corev1.LocalObjectReference{Name: secretName}
	}
	registryName := os.Getenv("PRESET_REGISTRY_NAME")
	imageName := formatRegistryImagePath(registryName, wObj.Tuning.Input.Image, "")
	return imageName, imagePullSecretRefs
}

func GetDataDestImageInfo(ctx context.Context, wObj *kaitov1alpha1.Workspace) (string, []corev1.LocalObjectReference) {
	registryName := os.Getenv("PRESET_REGISTRY_NAME")
	imageName := fmt.Sprintf("%s/%s", registryName, wObj.Tuning.Output.Image)
	imagePushSecretRefs := []corev1.LocalObjectReference{{Name: wObj.Tuning.Output.ImagePushSecret}}
	return imageName, imagePushSecretRefs
}

func CreatePresetConfigMap(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) error {
	// Copy Configmap from helm chart configmap into workspace
	releaseNamespace, err := utils.GetReleaseNamespace()
	if err != nil {
		return fmt.Errorf("failed to get release namespace: %v", err)
	}
	existingCM := &corev1.ConfigMap{}
	err = resources.GetResource(ctx, workspaceObj.Tuning.ConfigTemplate, workspaceObj.Namespace, kubeClient, existingCM)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		klog.Info("ConfigMap already exists in target namespace: %s, no action taken.\n", workspaceObj.Namespace)
		return nil
	}

	templateCM := &corev1.ConfigMap{}
	err = resources.GetResource(ctx, workspaceObj.Tuning.ConfigTemplate, releaseNamespace, kubeClient, templateCM)
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap from template namespace: %v", err)
	}

	templateCM.Namespace = workspaceObj.Namespace
	templateCM.ResourceVersion = "" // Clear metadata not needed for creation
	templateCM.UID = ""             // Clear UID

	// TODO: Any Custom Preset override logic for the configmap can go here
	err = resources.CreateResource(ctx, templateCM, kubeClient)
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap in target namespace, %s: %v", workspaceObj.Namespace, err)
	}

	return nil
}

func CreatePresetTuning(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) (client.Object, error) {
	// TODO
	return nil, nil
}
