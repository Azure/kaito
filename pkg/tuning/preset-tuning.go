package tuning

import (
	"context"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/model"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreatePresetTuning(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace,
	tuningObj *model.PresetParam, kubeClient client.Client) (client.Object, error) {
	// TODO

	// e.g. example from Inference
	//volume, volumeMount := configVolume(workspaceObj, inferenceObj)
	//commands, resourceReq := prepareInferenceParameters(ctx, inferenceObj)
	//image, imagePullSecrets := GetImageInfo(ctx, workspaceObj, inferenceObj)
	//
	//depObj = resources.GenerateDeploymentManifest(ctx, workspaceObj, image, imagePullSecrets, *workspaceObj.Resource.Count, commands,
	//	containerPorts, livenessProbe, readinessProbe, resourceReq, tolerations, volume, volumeMount)
	//
	//err := resources.CreateResource(ctx, depObj, kubeClient)
	//if client.IgnoreAlreadyExists(err) != nil {
	//	return nil, err
	//}
	//return depObj, nil

	return nil, nil
}
