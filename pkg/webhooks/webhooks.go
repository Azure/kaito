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

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
)

func NewWebhooks() []knativeinjection.ControllerConstructor {
	return []knativeinjection.ControllerConstructor{
		certificates.NewController,
		NewCRDValidationWebhook,
	}
}

func NewCRDValidationWebhook(ctx context.Context, _ configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,
		"validation.webhook.kdm.io",
		"/validate/workspace.kdm.io",
		Resources,
		func(ctx context.Context) context.Context { return ctx },
		true,
	)
}

var Resources = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	kdmv1alpha1.GroupVersion.WithKind("Workspace"): &kdmv1alpha1.Workspace{},
}
