package tuning

import (
	"context"
	"strings"

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

func CreatePresetTuning(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) (client.Object, error) {
	initContainers, imagePullSecrets, volumes, volumeMounts, err := prepareDataSource(ctx, workspaceObj, kubeClient)
	if err != nil {
		return nil, err
	}
	shmVolume, shmVolumeMount := utils.ConfigSHMVolume(workspaceObj)
	volumes = append(volumes, shmVolume)
	volumeMounts = append(volumeMounts, shmVolumeMount)

	commands, resourceReq := prepareTuningParameters(ctx, tuningObj)

	jobObj := resources.GenerateTuningJobManifest(ctx, workspaceObj, imagePullSecrets, *workspaceObj.Resource.Count, commands,
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
	case workspaceObj.Tuning.Input.HostPath != "":
		initContainers, volumes, volumeMounts = handleHostPathDataSource(ctx, workspaceObj)
	}
	return initContainers, imagePullSecrets, volumes, volumeMounts, nil
}

func getImageSecrets(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) []corev1.LocalObjectReference {
	var imagePullSecretRefs []corev1.LocalObjectReference
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

	volumes, volumeMounts := utils.ConfigDataVolume()
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
	volumes, volumeMounts := utils.ConfigDataVolume()
	return initContainers, volumes, volumeMounts
}

func handleHostPathDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) ([]corev1.Container, []corev1.Volume, []corev1.VolumeMount) {
	var initContainers []corev1.Container
	volumes, volumeMounts := utils.ConfigDataVolume()
	return initContainers, volumes, volumeMounts
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

// prepareTuningParameters builds a PyTorch command:
// accelerate launch <TORCH_PARAMS> baseCommand <MODEL_PARAMS>
// and sets the GPU resources required for inference.
// Returns the command and resource configuration.
func prepareTuningParameters(ctx context.Context, tuningObj *model.PresetParam) ([]string, corev1.ResourceRequirements) {
	torchCommand := utils.BuildCmdStr(tuningObj.BaseCommand, tuningObj.TorchRunParams)
	torchCommand = utils.BuildCmdStr(torchCommand, tuningObj.TorchRunRdzvParams)
	modelCommand := utils.BuildCmdStr(TuningFile, tuningObj.ModelRunParams)
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
