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
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetInstanceGPUCount(sku string) int {
	gpuConfig, exists := kaitov1alpha1.SupportedGPUConfigs[sku]
	if !exists {
		return 1
	}
	return gpuConfig.GPUCount
}

func GetTuningImageInfo(ctx context.Context, wObj *kaitov1alpha1.Workspace, presetObj *model.PresetParam) string {
	registryName := os.Getenv("PRESET_REGISTRY_NAME")
	return fmt.Sprintf("%s/%s:%s", registryName, "kaito-tuning-"+string(wObj.Tuning.Preset.Name), presetObj.Tag)
}

func GetDataSrcImageInfo(ctx context.Context, wObj *kaitov1alpha1.Workspace) (string, []corev1.LocalObjectReference) {
	imagePullSecretRefs := make([]corev1.LocalObjectReference, len(wObj.Tuning.Input.ImagePullSecrets))
	for i, secretName := range wObj.Tuning.Input.ImagePullSecrets {
		imagePullSecretRefs[i] = corev1.LocalObjectReference{Name: secretName}
	}
	return wObj.Tuning.Input.Image, imagePullSecretRefs
}

func GetDataDestImageInfo(ctx context.Context, wObj *kaitov1alpha1.Workspace) (string, []corev1.LocalObjectReference) {
	imagePushSecretRefs := []corev1.LocalObjectReference{{Name: wObj.Tuning.Output.ImagePushSecret}}
	return wObj.Tuning.Output.Image, imagePushSecretRefs
}

func EnsureTuningConfigMap(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
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
		klog.Infof("ConfigMap already exists in target namespace: %s, no action taken.\n", workspaceObj.Namespace)
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
