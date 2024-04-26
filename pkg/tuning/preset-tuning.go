package tuning

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
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

const (
	Port5000   = int32(5000)
	TuningFile = "fine_tuning_api.py"
)

var (
	containerPorts = []corev1.ContainerPort{{
		ContainerPort: Port5000,
	}}

	// Come up with valid liveness and readiness probes for fine-tuning
	// TODO: livenessProbe = &corev1.Probe{}
	// TODO: readinessProbe = &corev1.Probe{}

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
	initContainers, imagePullSecrets, volumes, volumeMounts, err := prepareDataSource(ctx, workspaceObj, kubeClient)
	if err != nil {
		return nil, err
	}

	err = EnsureTuningConfigMap(ctx, workspaceObj, tuningObj, kubeClient)
	if err != nil {
		return nil, err
	}

	shmVolume, shmVolumeMount := utils.ConfigSHMVolume(*workspaceObj.Resource.Count)
	if shmVolume.Name != "" {
		volumes = append(volumes, shmVolume)
	}
	if shmVolumeMount.Name != "" {
		volumeMounts = append(volumeMounts, shmVolumeMount)
	}

	cmVolume, cmVolumeMount := utils.ConfigCMVolume(workspaceObj.Tuning.ConfigTemplate)
	volumes = append(volumes, cmVolume)
	volumeMounts = append(volumeMounts, cmVolumeMount)

	modelCommand, err := prepareModelRunParameters(ctx, tuningObj)
	if err != nil {
		return nil, err
	}
	commands, resourceReq := prepareTuningParameters(ctx, workspaceObj, modelCommand, tuningObj)
	tuningImage := GetTuningImageInfo(ctx, workspaceObj, tuningObj)

	jobObj := resources.GenerateTuningJobManifest(ctx, workspaceObj, tuningImage, imagePullSecrets, *workspaceObj.Resource.Count, commands,
		containerPorts, nil, nil, resourceReq, tolerations, initContainers, volumes, volumeMounts)

	err = resources.CreateResource(ctx, jobObj, kubeClient)
	if client.IgnoreAlreadyExists(err) != nil {
		return nil, err
	}
	return jobObj, nil
}

// Now there are three options for DataSource: 1. URL - 2. HostPath - 3. Image
func prepareDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, kubeClient client.Client) ([]corev1.Container, []corev1.LocalObjectReference, []corev1.Volume, []corev1.VolumeMount, error) {
	var initContainers []corev1.Container
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	var imagePullSecrets []corev1.LocalObjectReference
	switch {
	case workspaceObj.Tuning.Input.Image != "":
		initContainers, volumes, volumeMounts = handleImageDataSource(ctx, workspaceObj)
		_, imagePullSecrets = GetDataSrcImageInfo(ctx, workspaceObj)
	}
	return initContainers, imagePullSecrets, volumes, volumeMounts, nil
}

func handleImageDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) ([]corev1.Container, []corev1.Volume, []corev1.VolumeMount) {
	var initContainers []corev1.Container
	// Constructing a multistep command that lists, copies, and then lists the destination
	command := "ls -la /data && cp -r /data/* " + utils.DefaultDataVolumePath + " && ls -la " + utils.DefaultDataVolumePath
	initContainers = append(initContainers, corev1.Container{
		Name:    "data-extractor",
		Image:   workspaceObj.Tuning.Input.Image,
		Command: []string{"sh", "-c", command},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data-volume",
				MountPath: utils.DefaultDataVolumePath,
			},
		},
	})

	volumes, volumeMounts := utils.ConfigDataVolume("")
	return initContainers, volumes, volumeMounts
}

func prepareModelRunParameters(ctx context.Context, tuningObj *model.PresetParam) (string, error) {
	modelCommand := utils.BuildCmdStr(TuningFile, tuningObj.ModelRunParams)
	return modelCommand, nil
}

// prepareTuningParameters builds a PyTorch command:
// accelerate launch <TORCH_PARAMS> baseCommand <MODEL_PARAMS>
// and sets the GPU resources required for tuning.
// Returns the command and resource configuration.
func prepareTuningParameters(ctx context.Context, wObj *kaitov1alpha1.Workspace, modelCommand string, tuningObj *model.PresetParam) ([]string, corev1.ResourceRequirements) {
	// Set # of processes to GPU Count
	numProcesses := getInstanceGPUCount(wObj.Resource.InstanceType)
	tuningObj.TorchRunParams["num_processes"] = fmt.Sprintf("%d", numProcesses)
	torchCommand := utils.BuildCmdStr(tuningObj.BaseCommand, tuningObj.TorchRunParams)
	torchCommand = utils.BuildCmdStr(torchCommand, tuningObj.TorchRunRdzvParams)
	commands := utils.ShellCmd(torchCommand + " " + modelCommand)

	resourceRequirements := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceName(resources.CapacityNvidiaGPU): resource.MustParse(tuningObj.GPUCountRequirement),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceName(resources.CapacityNvidiaGPU): resource.MustParse(tuningObj.GPUCountRequirement),
		},
	}

	return commands, resourceRequirements
}
