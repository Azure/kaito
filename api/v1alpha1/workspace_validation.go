package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"knative.dev/pkg/apis"
)

func (w *Workspace) SupportedVerbs() []admissionregistrationv1.OperationType {
	return []admissionregistrationv1.OperationType{
		admissionregistrationv1.Create,
		admissionregistrationv1.Update,
	}
}

func (w *Workspace) Validate(ctx context.Context) (errs *apis.FieldError) {
	klog.InfoS("Validating", "workspace", fmt.Sprintf("%s/%s", w.Namespace, w.Name))
	return errs
}
