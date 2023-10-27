// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package inference

import (
	"context"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/resources"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateTemplateInference(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, kubeClient client.Client) (client.Object, error) {
	depObj := resources.GenerateDeploymentManifestWithPodTemplate(ctx, workspaceObj)
	err := resources.CreateResource(ctx, client.Object(depObj), kubeClient)
	if client.IgnoreAlreadyExists(err) != nil {
		return nil, err
	}
	return depObj, nil
}
