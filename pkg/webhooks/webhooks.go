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

func NewWebhooks() []knativeinjection.ControllerConstructor {
	return []knativeinjection.ControllerConstructor{
		certificates.NewController,
		NewCRDValidationWebhook,
	}
}

func NewCRDValidationWebhook(ctx context.Context, _ configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,
		"validation.webhook.kaito.io",
		"/validate/workspace.kaito.io",
		Resources,
		func(ctx context.Context) context.Context { return ctx },
		true,
	)
}

var Resources = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	kaitov1alpha1.GroupVersion.WithKind("Workspace"): &kaitov1alpha1.Workspace{},
}
