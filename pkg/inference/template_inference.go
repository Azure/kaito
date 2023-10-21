package inference

import (
	"context"
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/resources"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateTemplateInference(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, kubeClient client.Client) error {
	klog.InfoS("CreateTemplateInference", "workspace", klog.KObj(workspaceObj))

	var depObj client.Object
	depObj = resources.GenerateDeploymentManifestWithPodTemplate(ctx, workspaceObj)
	err := resources.CreateResource(ctx, depObj, kubeClient)
	if client.IgnoreAlreadyExists(err) != nil {
		return err
	}

	if err := checkResourceStatus(depObj, kubeClient, 10*time.Minute); err != nil {
		return err
	}
	return nil
}
