// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

import (
	"context"
	"fmt"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/klog/v2"
	"knative.dev/pkg/apis"
)

func (w *RAGEngine) SupportedVerbs() []admissionregistrationv1.OperationType {
	return []admissionregistrationv1.OperationType{
		admissionregistrationv1.Create,
		admissionregistrationv1.Update,
	}
}

func (w *RAGEngine) Validate(ctx context.Context) (errs *apis.FieldError) {
	base := apis.GetBaseline(ctx)
	if base == nil {
		klog.InfoS("Validate creation", "ragengine", fmt.Sprintf("%s/%s", w.Namespace, w.Name))
		errs = errs.Also(w.validateCreate().ViaField("spec"))
	}
	return errs
}

func (w *RAGEngine) validateCreate() (errs *apis.FieldError) {
	if w.Spec.InferenceService == nil {
		errs = errs.Also(apis.ErrGeneric("InferenceService must be specified", ""))
	}
	return errs
}
