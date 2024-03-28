package tuning

import (
	"context"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/resources"
	"github.com/azure/kaito/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProbePath  = "/healthz"
	Port5000   = int32(5000)
	TuningFile = "fine_tuning_api.py"
)

func CreatePresetTuning(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) (client.Object, error) {
	volume, volumeMount := utils.ConfigVolume(workspaceObj)
	commands, resourceReq := prepareTuningParameters(ctx, tuningObj)

	getDataSource(ctx, workspaceObj, tuningObj, kubeClient)

	jobObj = resources.GenerateTuningJobManifest(ctx, workspaceObj /*TODO*/)
	//depObj = resources.GenerateDeploymentManifest(ctx, workspaceObj, image, imagePullSecrets, *workspaceObj.Resource.Count, commands,
	//	containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volume, volumeMount)

	err := resources.CreateResource(ctx, jobObj, kubeClient)
	if client.IgnoreAlreadyExists(err) != nil {
		return nil, err
	}
	return jobObj, nil
}

func getDataDestinationImage(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) (string, []corev1.LocalObjectReference) {
	imageName := workspaceObj.Tuning.Output.Image
	imagePushSecrets := []corev1.LocalObjectReference{{Name: workspaceObj.Tuning.Output.ImagePushSecret}}
	return imageName, imagePushSecrets
}

func getDataSourceImage(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) (string, []corev1.LocalObjectReference) {
	imagePullSecretRefs := []corev1.LocalObjectReference{}
	imageName := workspaceObj.Tuning.Input.Image
	for _, secretName := range workspaceObj.Tuning.Input.ImagePullSecrets {
		imagePullSecretRefs = append(imagePullSecretRefs, corev1.LocalObjectReference{Name: secretName})
	}
	return imageName, imagePullSecretRefs
}

// Now there are three options for DataSource: 1. URL - 2. HostPath - 3. Image
func getDataSource(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) (client.Object, error) {
	if workspaceObj.Tuning.Input.Image != "" {
		imageName, imagePullSecrets := getDataSourceImage(ctx, workspaceObj)
	} else if len(workspaceObj.Tuning.Input.URLs) > 0 {
		// TODO
	} else if workspaceObj.Tuning.Input.HostPath != "" {
		// TODO
	}
	return nil, nil
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
