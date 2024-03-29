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
	return nil, nil
}
