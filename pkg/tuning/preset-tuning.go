package tuning

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/resources"
	"github.com/azure/kaito/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func GetTuningImageInfo(ctx context.Context, wObj *kaitov1alpha1.Workspace, presetObj *model.PresetParam) string {
	// Only support public presets for tuning for now
	imageName := string(wObj.Tuning.Preset.Name)
	imageTag := presetObj.Tag
	registryName := os.Getenv("PRESET_REGISTRY_NAME")
	imageName = fmt.Sprintf("%s/kaito-tuning-%s:%s", registryName, imageName, imageTag)
	return imageName
}

func GetDataDestImageInfo(ctx context.Context, wObj *kaitov1alpha1.Workspace) (string, []corev1.LocalObjectReference) {
	imagePushSecretRefs := []corev1.LocalObjectReference{}
	imagePushSecretRefs = append(imagePushSecretRefs, corev1.LocalObjectReference{Name: wObj.Tuning.Output.ImagePushSecret})
	registryName := os.Getenv("PRESET_REGISTRY_NAME")
	imageName := fmt.Sprintf("%s/%s", registryName, wObj.Tuning.Output.Image)
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
	initContainers, imagePullSecrets, volumes, volumeMounts, err := prepareDataSource(ctx, workspaceObj, kubeClient)
	if err != nil {
		return nil, err
	}

	err = CreatePresetConfigMap(ctx, workspaceObj, tuningObj, kubeClient)
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
		containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, initContainers, volumes, volumeMounts)

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
		imagePullSecrets = getImageSecrets(ctx, workspaceObj)
	case len(workspaceObj.Tuning.Input.URLs) > 0:
		initContainers, volumes, volumeMounts = handleURLDataSource(ctx, workspaceObj)
	case workspaceObj.Tuning.Input.Volume != nil:
		initContainers, volumes, volumeMounts = handleVolumeDataSource(ctx, workspaceObj)
	}
	return initContainers, imagePullSecrets, volumes, volumeMounts, nil
}

func getImageSecrets(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) []corev1.LocalObjectReference {
	imagePullSecretRefs := []corev1.LocalObjectReference{}
	for _, secretName := range workspaceObj.Tuning.Input.ImagePullSecrets {
		imagePullSecretRefs = append(imagePullSecretRefs, corev1.LocalObjectReference{Name: secretName})
	}
	return imagePullSecretRefs
}

func handleImageDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) ([]corev1.Container, []corev1.Volume, []corev1.VolumeMount) {
	var initContainers []corev1.Container
	initContainers = append(initContainers, corev1.Container{
		Name:    "data-extractor",
		Image:   workspaceObj.Tuning.Input.Image,
		Command: []string{"sh", "-c", "your-extraction-script.sh"}, // Your script to extract data
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data-volume",
				MountPath: "/data",
			},
		},
	})

	volumes, volumeMounts := utils.ConfigDataVolume("")
	return initContainers, volumes, volumeMounts
}

func handleURLDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) ([]corev1.Container, []corev1.Volume, []corev1.VolumeMount) {
	var initContainers []corev1.Container
	// TODO: Fix up init container placeholders
	initContainers = append(initContainers, corev1.Container{
		Name:    "data-downloader",
		Image:   "appropriate-image-for-downloading",
		Command: []string{"sh", "-c", "download-script.sh"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data-volume",
				MountPath: "/data",
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "DATA_URLS",
				Value: strings.Join(workspaceObj.Tuning.Input.URLs, " "),
			},
		},
	})
	volumes, volumeMounts := utils.ConfigDataVolume("")
	return initContainers, volumes, volumeMounts
}

func handleVolumeDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) ([]corev1.Container, []corev1.Volume, []corev1.VolumeMount) {
	var initContainers []corev1.Container
	_ = workspaceObj.Tuning.Input.Volume
	// TODO
	// volumes, volumeMounts := utils.ConfigDataVolume(hostPath)
	// return initContainers, volumes, volumeMounts
	return initContainers, []corev1.Volume{}, []corev1.VolumeMount{}
}

func getDataDestinationImage(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) (string, []corev1.LocalObjectReference) {
	imageName := workspaceObj.Tuning.Output.Image
	imagePushSecrets := []corev1.LocalObjectReference{{Name: workspaceObj.Tuning.Output.ImagePushSecret}}
	return imageName, imagePushSecrets
}

func getDataDestination(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) (client.Object, error) {
	// TODO
	return nil, nil
}

func getInstanceGPUCount(sku string) int {
	gpuConfig, exists := kaitov1alpha1.SupportedGPUConfigs[sku]
	if !exists {
		return 1
	}
	return gpuConfig.GPUCount
}

func prepareModelRunParameters(ctx context.Context, tuningObj *model.PresetParam) (string, error) {
	modelCommand := utils.BuildCmdStr(TuningFile, tuningObj.ModelRunParams)
	return modelCommand, nil
}

// prepareTuningParameters builds a PyTorch command:
// accelerate launch <TORCH_PARAMS> baseCommand <MODEL_PARAMS>
// and sets the GPU resources required for inference.
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
