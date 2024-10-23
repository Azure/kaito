// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package webhooks

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	knativeinjection "knative.dev/pkg/injection"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/validation"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
)

func NewWorkspaceWebhooks() []knativeinjection.ControllerConstructor {
	return []knativeinjection.ControllerConstructor{
		certificates.NewController,
		NewWorkspaceCRDValidationWebhook,
	}
}

func NewWorkspaceCRDValidationWebhook(ctx context.Context, _ configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,
		"validation.workspace.kaito.sh",
		"/validate/workspace.kaito.sh",
		WorkspaceResources,
		func(ctx context.Context) context.Context { return ctx },
		true,
	)
}

var WorkspaceResources = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	kaitov1alpha1.GroupVersion.WithKind("Workspace"): &kaitov1alpha1.Workspace{},
}

func NewRAGEngineWebhooks() []knativeinjection.ControllerConstructor {
	return []knativeinjection.ControllerConstructor{
		certificates.NewController,
		NewRAGEngineCRDValidationWebhook,
	}
}

func NewRAGEngineCRDValidationWebhook(ctx context.Context, _ configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,
		"validation.ragengine.kaito.sh",
		"/validate/ragengine.kaito.sh",
		RAGEngineResources,
		func(ctx context.Context) context.Context { return ctx },
		true,
	)
}

var RAGEngineResources = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	kaitov1alpha1.GroupVersion.WithKind("RAGEngine"): &kaitov1alpha1.RAGEngine{},
}
