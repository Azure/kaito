package tuning

import (
	"context"
	"fmt"
	"k8s.io/utils/pointer"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

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

func EnsureTuningConfigMap(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) error {
	// Copy Configmap from helm chart configmap into workspace
	existingCM := &corev1.ConfigMap{}
	err := resources.GetResource(ctx, workspaceObj.Tuning.ConfigTemplate, workspaceObj.Namespace, kubeClient, existingCM)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		klog.Infof("ConfigMap already exists in target namespace: %s, no action taken.\n", workspaceObj.Namespace)
		return nil
	}

	releaseNamespace, err := utils.GetReleaseNamespace()
	if err != nil {
		return fmt.Errorf("failed to get release namespace: %v", err)
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

func dockerSidecarScriptPushImage(image string) string {
	// TODO: Override output path if specified in trainingconfig (instead of /mnt/results)
	return fmt.Sprintf(`
# Start the Docker daemon in the background with specific options for DinD
dockerd &
# Wait for the Docker daemon to be ready
while ! docker info > /dev/null 2>&1; do
  echo "Waiting for Docker daemon to start..."
  sleep 1
done
echo 'Docker daemon started'

while true; do
  FILE_PATH=$(find /mnt/results -name 'fine_tuning_completed.txt')
  if [ ! -z "$FILE_PATH" ]; then
    echo "FOUND TRAINING COMPLETED FILE at $FILE_PATH"

    PARENT_DIR=$(dirname "$FILE_PATH")
    echo "Parent directory is $PARENT_DIR"

    TEMP_CONTEXT=$(mktemp -d)
    cp "$PARENT_DIR/adapter_config.json" "$TEMP_CONTEXT/adapter_config.json"
    cp -r "$PARENT_DIR/adapter_model.safetensors" "$TEMP_CONTEXT/adapter_model.safetensors"

    # Create a minimal Dockerfile
    echo 'FROM scratch
    ADD adapter_config.json /
    ADD adapter_model.safetensors /' > "$TEMP_CONTEXT/Dockerfile"

    docker build -t %s "$TEMP_CONTEXT"
    docker push %s

    # Cleanup: Remove the temporary directory
    rm -rf "$TEMP_CONTEXT"

    # Remove the file to prevent repeated builds
    rm "$FILE_PATH"
    echo "Upload complete"
    exit 0
  fi
  sleep 10  # Check every 10 seconds
done`, image, image)
}

func setupDefaultSharedVolumes(workspaceObj *kaitov1alpha1.Workspace) ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	// Add shared volume for shared memory (multi-node)
	shmVolume, shmVolumeMount := utils.ConfigSHMVolume(*workspaceObj.Resource.Count)
	if shmVolume.Name != "" {
		volumes = append(volumes, shmVolume)
	}
	if shmVolumeMount.Name != "" {
		volumeMounts = append(volumeMounts, shmVolumeMount)
	}

	// Add shared volume for tuning parameters
	cmVolume, cmVolumeMount := utils.ConfigCMVolume(workspaceObj.Tuning.ConfigTemplate)
	volumes = append(volumes, cmVolume)
	volumeMounts = append(volumeMounts, cmVolumeMount)

	// Add shared volume for results dir
	resultsVolume, resultsVolumeMount := utils.ConfigResultsVolume()
	if resultsVolume.Name != "" {
		volumes = append(volumes, resultsVolume)
	}
	if resultsVolumeMount.Name != "" {
		volumeMounts = append(volumeMounts, resultsVolumeMount)
	}
	return volumes, volumeMounts
}

func CreatePresetTuning(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) (client.Object, error) {
	var initContainers, sidecarContainers []corev1.Container
	volumes, volumeMounts := setupDefaultSharedVolumes(workspaceObj)

	initContainer, imagePullSecrets, dataSourceVolume, dataSourceVolumeMount, err := prepareDataSource(ctx, workspaceObj)
	if err != nil {
		return nil, err
	}
	volumes = append(volumes, dataSourceVolume)
	volumeMounts = append(volumeMounts, dataSourceVolumeMount)
	if initContainer.Name != "" {
		initContainers = append(initContainers, *initContainer)
	}

	sidecarContainer, imagePushSecret, dataDestVolume, dataDestVolumeMount, err := prepareDataDestination(ctx, workspaceObj)
	if err != nil {
		return nil, err
	}
	volumes = append(volumes, dataDestVolume)
	volumeMounts = append(volumeMounts, dataDestVolumeMount)
	if sidecarContainer != nil {
		sidecarContainers = append(sidecarContainers, *sidecarContainer)
	}
	if imagePushSecret != nil {
		imagePullSecrets = append(imagePullSecrets, *imagePushSecret)
	}

	err = EnsureTuningConfigMap(ctx, workspaceObj, tuningObj, kubeClient)
	if err != nil {
		return nil, err
	}

	modelCommand, err := prepareModelRunParameters(ctx, tuningObj)
	if err != nil {
		return nil, err
	}
	commands, resourceReq := prepareTuningParameters(ctx, workspaceObj, modelCommand, tuningObj)
	tuningImage := GetTuningImageInfo(ctx, workspaceObj, tuningObj)

	jobObj := resources.GenerateTuningJobManifest(ctx, workspaceObj, tuningImage, imagePullSecrets, *workspaceObj.Resource.Count, commands,
		containerPorts, nil, nil, resourceReq, tolerations, initContainers, sidecarContainers, volumes, volumeMounts)

	err = resources.CreateResource(ctx, jobObj, kubeClient)
	if client.IgnoreAlreadyExists(err) != nil {
		return nil, err
	}
	return jobObj, nil
}

// Now there are two options for data destination 1. HostPath - 2. Image
func prepareDataDestination(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) (*corev1.Container, *corev1.LocalObjectReference, corev1.Volume, corev1.VolumeMount, error) {
	var sidecarContainer *corev1.Container
	var volume corev1.Volume
	var volumeMount corev1.VolumeMount
	var imagePushSecret *corev1.LocalObjectReference
	switch {
	case workspaceObj.Tuning.Output.Image != "":
		image, secret := workspaceObj.Tuning.Output.Image, workspaceObj.Tuning.Output.ImagePushSecret
		imagePushSecret = &corev1.LocalObjectReference{Name: secret}
		sidecarContainer, volume, volumeMount = handleImageDataDestination(ctx, image, secret)
		// TODO: Future PR include
		//case workspaceObj.Tuning.Output.Volume != nil:
	}
	return sidecarContainer, imagePushSecret, volume, volumeMount, nil
}

func handleImageDataDestination(ctx context.Context, image, imagePushSecret string) (*corev1.Container, corev1.Volume, corev1.VolumeMount) {
	sidecarContainer := &corev1.Container{
		Name:  "docker-sidecar",
		Image: "docker:dind",
		SecurityContext: &corev1.SecurityContext{
			Privileged: pointer.BoolPtr(true),
		},
		Command: []string{"/bin/sh", "-c"},
		Args:    []string{dockerSidecarScriptPushImage(image)},
	}

	volume, volumeMount := utils.ConfigImagePushSecretVolume(imagePushSecret)
	return sidecarContainer, volume, volumeMount
}

// Now there are three options for DataSource: 1. URL - 2. HostPath - 3. Image
func prepareDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) (*corev1.Container, []corev1.LocalObjectReference, corev1.Volume, corev1.VolumeMount, error) {
	var initContainer *corev1.Container
	var volume corev1.Volume
	var volumeMount corev1.VolumeMount
	var imagePullSecrets []corev1.LocalObjectReference
	switch {
	case workspaceObj.Tuning.Input.Image != "":
		var image string
		image, imagePullSecrets = GetDataSrcImageInfo(ctx, workspaceObj)
		initContainer, volume, volumeMount = handleImageDataSource(ctx, image)
	case len(workspaceObj.Tuning.Input.URLs) > 0:
		initContainer, volume, volumeMount = handleURLDataSource(ctx, workspaceObj)
		// TODO: Future PR include
		// case workspaceObj.Tuning.Input.Volume != nil:
	}
	return initContainer, imagePullSecrets, volume, volumeMount, nil
}

func handleImageDataSource(ctx context.Context, image string) (*corev1.Container, corev1.Volume, corev1.VolumeMount) {
	// Constructing a multistep command that lists, copies, and then lists the destination
	command := "ls -la /data && cp -r /data/* " + utils.DefaultDataVolumePath + " && ls -la " + utils.DefaultDataVolumePath
	initContainer := &corev1.Container{
		Name:    "data-extractor",
		Image:   image,
		Command: []string{"sh", "-c", command},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data-volume",
				MountPath: utils.DefaultDataVolumePath,
			},
		},
	}

	volume, volumeMount := utils.ConfigDataVolume(nil)
	return initContainer, volume, volumeMount
}

func handleURLDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) (*corev1.Container, corev1.Volume, corev1.VolumeMount) {
	initContainer := &corev1.Container{
		Name:  "data-downloader",
		Image: "curlimages/curl",
		Command: []string{"sh", "-c", `
			for url in $DATA_URLS; do
				filename=$(basename "$url" | sed 's/[?=&]/_/g')
				curl -sSL $url -o $DATA_VOLUME_PATH/$filename
			done
		`},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data-volume",
				MountPath: utils.DefaultDataVolumePath,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "DATA_URLS",
				Value: strings.Join(workspaceObj.Tuning.Input.URLs, " "),
			},
			{
				Name:  "DATA_VOLUME_PATH",
				Value: utils.DefaultDataVolumePath,
			},
		},
	}
	volume, volumeMount := utils.ConfigDataVolume(nil)
	return initContainer, volume, volumeMount
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
	if tuningObj.TorchRunParams == nil {
		tuningObj.TorchRunParams = make(map[string]string)
	}
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
