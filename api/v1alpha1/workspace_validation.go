// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

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
	N_SERIES_PREFIX = "Standard_N"
	D_SERIES_PREFIX = "Standard_D"
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
		errs = errs.Also(
			w.Inference.validateCreate().ViaField("inference"),
			w.Resource.validateCreate(w.Inference).ViaField("resource"),
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

func (r *ResourceSpec) validateCreate(inference InferenceSpec) (errs *apis.FieldError) {
	var presetName string
	if inference.Preset != nil {
		presetName = strings.ToLower(string(inference.Preset.Name))
	}
	instanceType := string(r.InstanceType)

	// Check if instancetype exists in our SKUs map
	if skuConfig, exists := SupportedGPUConfigs[instanceType]; exists {
		if inference.Preset != nil {
			// Validate GPU count for given SKU
			if presetReq, ok := PresetRequirementsMap[presetName]; ok {
				machineCount := *r.Count
				totalNumGPUs := machineCount * skuConfig.GPUCount
				totalGPUMem := machineCount * skuConfig.GPUMem * skuConfig.GPUCount

				// Separate the checks for specific error messages
				if totalNumGPUs < presetReq.MinGPUCount {
					errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Insufficient number of GPUs: Instance type %s provides %d, but preset %s requires at least %d", instanceType, totalNumGPUs, presetName, presetReq.MinGPUCount), "instanceType"))
				}
				if skuConfig.GPUMem < presetReq.MinMemoryPerGPU {
					errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Insufficient GPU memory: Instance type %s provides %d per GPU, but preset %s requires at least %d per GPU", instanceType, skuConfig.GPUMem, presetName, presetReq.MinMemoryPerGPU), "instanceType"))
				}
				if totalGPUMem < presetReq.MinTotalMemory {
					errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Insufficient total GPU memory: Instance type %s has a total of %d, but preset %s requires at least %d", instanceType, totalGPUMem, presetName, presetReq.MinTotalMemory), "instanceType"))
				}
			} else {
				errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Unsupported preset name %s", presetName), "presetName"))
			}
		}
	} else {
		// Check for other instancetypes pattern matches
		if !strings.HasPrefix(instanceType, N_SERIES_PREFIX) && !strings.HasPrefix(instanceType, D_SERIES_PREFIX) {
			errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Unsupported instance type %s. Supported SKUs: %s", instanceType, getSupportedSKUs()), "instanceType"))
		}
	}

	// Validate labelSelector
	if _, err := metav1.LabelSelectorAsMap(r.LabelSelector); err != nil {
		errs = errs.Also(apis.ErrInvalidValue(err.Error(), "labelSelector"))
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
	// Check if both Preset and Template are not set
	if i.Preset == nil && i.Template == nil {
		errs = errs.Also(apis.ErrMissingField("Preset or Template must be specified"))
	}

	// Check if both Preset and Template are set at the same time
	if i.Preset != nil && i.Template != nil {
		errs = errs.Also(apis.ErrGeneric("Preset and Template cannot be set at the same time"))
	}

	var presetName string
	if i.Preset != nil {
		presetName = strings.ToLower(string(i.Preset.Name))
		// Validate preset name
		if !isValidPreset(presetName) {
			errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Unsupported preset name %s", presetName), "presetName"))
		}
		// Additional validations for Preset
		if i.Preset.PresetMeta.AccessMode == "private" && i.Preset.PresetOptions.Image == "" {
			errs = errs.Also(apis.ErrGeneric("When AccessMode is private, an image must be provided in PresetOptions"))
		}
		// Note: we don't enforce private access mode to have image secrets, in case anonymous pulling is enabled
	}
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
