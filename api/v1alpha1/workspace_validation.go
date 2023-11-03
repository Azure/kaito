package v1alpha1

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"knative.dev/pkg/apis"
)

const (
	N_SERIES_PREFIX = "standard_n"
	D_SERIES_PREFIX = "standard_d"
)

func (w *Workspace) SupportedVerbs() []admissionregistrationv1.OperationType {
	return []admissionregistrationv1.OperationType{
		admissionregistrationv1.Create,
		admissionregistrationv1.Update,
	}
}

func (w *Workspace) Validate(ctx context.Context) (errs *apis.FieldError) {

	base := apis.GetBaseline(ctx)
	if base == nil {
		klog.InfoS("Validate creation", "workspace", fmt.Sprintf("%s/%s", w.Namespace, w.Name))
		/* TODO
		Add the following checks when creating a new workspace:
		- For instancetype, check againt a hardcode list with existing GPU skus that we support. For these skus,
		  validate if the total GPU count is sufficient to run the preset workload if any.
		- For other instancetypes, do pattern matches and for N, D series, we let it pass. Otherwise, fail the check.
		- For labelSelector, call metav1.LabelSelectorAsMap. If the method returns error, meaning unsupported expressions are found, fail the check.
		- The preset name needs to be supported enum.
		*/

		// Validate preset name
		if !isValidPreset(string(w.Inference.Preset.Name)) {
			errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Unsupported preset name %s", w.Inference.Preset.Name), "presetName"))
		}

		// Check if instancetype exists in our SKUs map
		if sku, exists := SupportedGPUConfigs[w.Resource.InstanceType]; exists {
			// Validate GPU count for given SKU (if exists)
			if count, ok := sku.Counts[string(w.Inference.Preset.Name)]; ok {
				if *w.Resource.Count < count {
					errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Insufficient machine count for SKU %s", w.Resource.InstanceType), "instanceType"))
				}
			}
		} else {
			// Check for other instancetypes pattern matches
			if !strings.HasPrefix(w.Resource.InstanceType, N_SERIES_PREFIX) && !strings.HasPrefix(w.Resource.InstanceType, D_SERIES_PREFIX) {
				errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Unsupported instance type %s", w.Resource.InstanceType), "instanceType"))
			}
		}

		// Validate labelSelector
		if _, err := metav1.LabelSelectorAsMap(w.Resource.LabelSelector); err != nil {
			errs = errs.Also(apis.ErrInvalidValue(err.Error(), "labelSelector"))
		}

		errs = errs.Also(
			w.Inference.validateCreate().ViaField("inference"),
		)
	} else {
		klog.InfoS("Validate update", "workspace", fmt.Sprintf("%s/%s", w.Namespace, w.Name))
		old := base.(*Workspace)
		errs = errs.Also(
			w.Resource.validateUpdate(&old.Resource).ViaField("resource"),
			w.Inference.validateUpdate(&old.Inference).ViaField("inference"),
		)
	}
	return errs
}

func (r *ResourceSpec) validateUpdate(old *ResourceSpec) (errs *apis.FieldError) {
	// We disable changing node count for now.
	if r.Count != nil && old.Count != nil && *r.Count != *old.Count {
		errs = errs.Also(apis.ErrGeneric("field is immutable", "count"))
	}
	if r.InstanceType != old.InstanceType {
		errs = errs.Also(apis.ErrGeneric("field is immutable", "instanceType"))
	}
	newLabels, err0 := metav1.LabelSelectorAsMap(r.LabelSelector)
	oldLabels, err1 := metav1.LabelSelectorAsMap(old.LabelSelector)
	if err0 != nil || err1 != nil {
		errs = errs.Also(apis.ErrGeneric("Only allow matchLabels or 'IN' matchExpression", "labelSelector"))
	} else {
		if !reflect.DeepEqual(newLabels, oldLabels) {
			errs = errs.Also(apis.ErrGeneric("field is immutable", "labelSelector"))
		}
	}
	return errs
}

func (i *InferenceSpec) validateCreate() (errs *apis.FieldError) {
	if i.Preset != nil && i.Template != nil {
		errs = errs.Also(apis.ErrGeneric("preset and template cannot be set at the same time"))
	}
	if i.Preset != nil && i.Preset.PresetMeta.AccessMode == "private" && i.Preset.PresetOptions.Image == "" {
		errs = errs.Also(apis.ErrGeneric("When AccessMode is private, an image must be provided in PresetOptions"))
	}
	// Note: we don't enforce private access mode to have image secrets, incase anonymous pulling is enabled
	return errs
}

func (i *InferenceSpec) validateUpdate(old *InferenceSpec) (errs *apis.FieldError) {
	if !reflect.DeepEqual(i.Preset, old.Preset) {
		errs = errs.Also(apis.ErrGeneric("field is immutable", "preset"))
	}
	//inference.template can be changed, but cannot be unset.
	if (i.Template != nil && old.Template == nil) || (i.Template == nil && old.Template != nil) {
		errs = errs.Also(apis.ErrGeneric("field cannot be unset/set if it was set/unset", "template"))
	}

	return errs
}
