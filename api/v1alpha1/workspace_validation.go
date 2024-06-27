// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/azure/kaito/pkg/utils"
	"github.com/azure/kaito/pkg/utils/plugin"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

const (
	N_SERIES_PREFIX = "Standard_N"
	D_SERIES_PREFIX = "Standard_D"

	DefaultLoraConfigMapTemplate  = "lora-params-template"
	DefaultQloraConfigMapTemplate = "qlora-params-template"
	MaxAdaptersNumber             = 10
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
		errs = errs.Also(w.validateCreate().ViaField("spec"))
		if w.Inference != nil {
			// TODO: Add Adapter Spec Validation - Including DataSource Validation for Adapter
			errs = errs.Also(w.Resource.validateCreateWithInference(w.Inference).ViaField("resource"),
				w.Inference.validateCreate().ViaField("inference"))
		}
		if w.Tuning != nil {
			// TODO: Add validate resource based on Tuning Spec
			errs = errs.Also(w.Resource.validateCreateWithTuning(w.Tuning).ViaField("resource"),
				w.Tuning.validateCreate(ctx, w.Namespace).ViaField("tuning"))
		}
	} else {
		klog.InfoS("Validate update", "workspace", fmt.Sprintf("%s/%s", w.Namespace, w.Name))
		old := base.(*Workspace)
		errs = errs.Also(
			w.validateUpdate(old).ViaField("spec"),
			w.Resource.validateUpdate(&old.Resource).ViaField("resource"),
		)
		if w.Inference != nil {
			errs = errs.Also(w.Inference.validateUpdate(old.Inference).ViaField("inference"))
		}
		if w.Tuning != nil {
			errs = errs.Also(w.Tuning.validateUpdate(old.Tuning).ViaField("tuning"))
		}
	}
	return errs
}

func (w *Workspace) validateCreate() (errs *apis.FieldError) {
	if w.Inference == nil && w.Tuning == nil {
		errs = errs.Also(apis.ErrGeneric("Either Inference or Tuning must be specified, not neither", ""))
	}
	if w.Inference != nil && w.Tuning != nil {
		errs = errs.Also(apis.ErrGeneric("Either Inference or Tuning must be specified, but not both", ""))
	}
	return errs
}

func (w *Workspace) validateUpdate(old *Workspace) (errs *apis.FieldError) {
	if (old.Inference == nil && w.Inference != nil) || (old.Inference != nil && w.Inference == nil) {
		errs = errs.Also(apis.ErrGeneric("Inference field cannot be toggled once set", "inference"))
	}

	if (old.Tuning == nil && w.Tuning != nil) || (old.Tuning != nil && w.Tuning == nil) {
		errs = errs.Also(apis.ErrGeneric("Tuning field cannot be toggled once set", "tuning"))
	}
	return errs
}

func ValidateDNSSubdomain(name string) bool {
	var dnsSubDomainRegexp = regexp.MustCompile(`^(?i:[a-z0-9]([-a-z0-9]*[a-z0-9])?)$`)
	if len(name) < 1 || len(name) > 253 {
		return false
	}
	return dnsSubDomainRegexp.MatchString(name)
}

func (r *AdapterSpec) validateCreateorUpdate() (errs *apis.FieldError) {
	if r.Source == nil {
		errs = errs.Also(apis.ErrMissingField("Source"))
	} else {
		errs = errs.Also(r.Source.validateCreate().ViaField("Adapters"))

		if r.Source.Name == "" {
			errs = errs.Also(apis.ErrMissingField("Name of Adapter field must be specified"))
		} else if !ValidateDNSSubdomain(r.Source.Name) {
			errs = errs.Also(apis.ErrMissingField("Name of Adapter must be a valid DNS subdomain value"))
		}
		if r.Source.Image == "" {
			errs = errs.Also(apis.ErrMissingField("Image of Adapter field must be specified"))
		}
		if r.Strength == nil {
			var defaultStrength = "1.0"
			r.Strength = &defaultStrength
		}
		strength, err := strconv.ParseFloat(*r.Strength, 64)
		if err != nil {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Invalid strength value for Adapter '%s': %v", r.Source.Name, err), "adapter"))
		}
		if strength < 0 || strength > 1.0 {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Strength value for Adapter '%s' must be between 0 and 1", r.Source.Name), "adapter"))
		}

	}
	return errs
}

func (r *TuningSpec) validateCreate(ctx context.Context, workspaceNamespace string) (errs *apis.FieldError) {
	methodLowerCase := strings.ToLower(string(r.Method))
	if methodLowerCase != string(TuningMethodLora) && methodLowerCase != string(TuningMethodQLora) {
		errs = errs.Also(apis.ErrInvalidValue(r.Method, "Method"))
	}
	if r.Config == "" {
		klog.InfoS("Tuning config not specified. Using default based on method.")
		releaseNamespace, err := utils.GetReleaseNamespace()
		if err != nil {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Failed to determine release namespace: %v", err), "namespace"))
		}
		defaultConfigMapTemplateName := ""
		if methodLowerCase == string(TuningMethodLora) {
			defaultConfigMapTemplateName = DefaultLoraConfigMapTemplate
		} else if methodLowerCase == string(TuningMethodQLora) {
			defaultConfigMapTemplateName = DefaultQloraConfigMapTemplate
		}
		if err := r.validateConfigMap(ctx, releaseNamespace, methodLowerCase, defaultConfigMapTemplateName); err != nil {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Failed to evaluate validateConfigMap: %v", err), "Config"))
		}
	} else {
		if err := r.validateConfigMap(ctx, workspaceNamespace, methodLowerCase, r.Config); err != nil {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Failed to evaluate validateConfigMap: %v", err), "Config"))
		}
	}
	if r.Input == nil {
		errs = errs.Also(apis.ErrMissingField("Input"))
	} else {
		errs = errs.Also(r.Input.validateCreate().ViaField("Input"))
	}
	if r.Output == nil {
		errs = errs.Also(apis.ErrMissingField("Output"))
	} else {
		errs = errs.Also(r.Output.validateCreate().ViaField("Output"))
	}
	// Currently require a preset to specified, in future we can consider defining a template
	if r.Preset == nil {
		errs = errs.Also(apis.ErrMissingField("Preset"))
	} else if presetName := string(r.Preset.Name); !isValidPreset(presetName) {
		errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Unsupported tuning preset name %s", presetName), "presetName"))
	}
	return errs
}

func (r *TuningSpec) validateUpdate(old *TuningSpec) (errs *apis.FieldError) {
	if r.Input == nil {
		errs = errs.Also(apis.ErrMissingField("Input"))
	} else {
		errs = errs.Also(r.Input.validateUpdate(old.Input, true).ViaField("Input"))
	}
	if r.Output == nil {
		errs = errs.Also(apis.ErrMissingField("Output"))
	} else {
		errs = errs.Also(r.Output.validateUpdate(old.Output).ViaField("Output"))
	}
	if !reflect.DeepEqual(old.Preset, r.Preset) {
		errs = errs.Also(apis.ErrGeneric("Preset cannot be changed", "Preset"))
	}
	oldMethod, newMethod := strings.ToLower(string(old.Method)), strings.ToLower(string(r.Method))
	if !reflect.DeepEqual(oldMethod, newMethod) {
		errs = errs.Also(apis.ErrGeneric("Method cannot be changed", "Method"))
	}
	// Consider supporting config fields changing
	return errs
}

func (r *DataSource) validateCreate() (errs *apis.FieldError) {
	sourcesSpecified := 0
	if len(r.URLs) > 0 {
		sourcesSpecified++
	}
	if r.Volume != nil {
		errs = errs.Also(apis.ErrInvalidValue("Volume support is not implemented yet", "Volume"))
		sourcesSpecified++
	}
	// Regex checks for a / and a colon followed by a tag
	if r.Image != "" {
		re := regexp.MustCompile(`^(.+/[^:/]+):([^:/]+)$`)
		if !re.MatchString(r.Image) {
			errs = errs.Also(apis.ErrInvalidValue("Invalid image format, require full input image URL", "Image"))
		}
		sourcesSpecified++
	}

	// Ensure exactly one of URLs, Volume, or Image is specified
	if sourcesSpecified != 1 {
		errs = errs.Also(apis.ErrGeneric("Exactly one of URLs, Volume, or Image must be specified", "URLs", "Volume", "Image"))
	}

	return errs
}

func (r *DataSource) validateUpdate(old *DataSource, isTuning bool) (errs *apis.FieldError) {
	if isTuning && !reflect.DeepEqual(old.Name, r.Name) {
		errs = errs.Also(apis.ErrInvalidValue("During tuning Name field cannot be changed once set", "Name"))
	}
	if r.Volume != nil {
		errs = errs.Also(apis.ErrInvalidValue("Volume support is not implemented yet", "Volume"))
	}
	oldURLs := make([]string, len(old.URLs))
	copy(oldURLs, old.URLs)
	sort.Strings(oldURLs)

	newURLs := make([]string, len(r.URLs))
	copy(newURLs, r.URLs)
	sort.Strings(newURLs)

	if !reflect.DeepEqual(oldURLs, newURLs) {
		errs = errs.Also(apis.ErrInvalidValue("URLs field cannot be changed once set", "URLs"))
	}
	// TODO: check if the Volume is changed
	if old.Image != r.Image {
		errs = errs.Also(apis.ErrInvalidValue("Image field cannot be changed once set", "Image"))
	}

	oldSecrets := make([]string, len(old.ImagePullSecrets))
	copy(oldSecrets, old.ImagePullSecrets)
	sort.Strings(oldSecrets)

	newSecrets := make([]string, len(r.ImagePullSecrets))
	copy(newSecrets, r.ImagePullSecrets)
	sort.Strings(newSecrets)

	if !reflect.DeepEqual(oldSecrets, newSecrets) {
		errs = errs.Also(apis.ErrInvalidValue("ImagePullSecrets field cannot be changed once set", "ImagePullSecrets"))
	}
	return errs
}

func (r *DataDestination) validateCreate() (errs *apis.FieldError) {
	destinationsSpecified := 0
	// TODO: Implement Volumes
	if r.Volume != nil {
		errs = errs.Also(apis.ErrInvalidValue("Volume support is not implemented yet", "Volume"))
		destinationsSpecified++
	}
	if r.Image != "" {
		// Regex checks for a / and a colon followed by a tag
		re := regexp.MustCompile(`^(.+/[^:/]+):([^:/]+)$`)
		if !re.MatchString(r.Image) {
			errs = errs.Also(apis.ErrInvalidValue("Invalid image format, require full output image URL", "Image"))
		}
		// Cloud Provider requires credentials to push image
		if r.ImagePushSecret == "" {
			errs = errs.Also(apis.ErrMissingField("Must specify imagePushSecret with destination image"))
		}
		destinationsSpecified++
	}

	// If no destination is specified, return an error
	if destinationsSpecified == 0 {
		errs = errs.Also(apis.ErrMissingField("At least one of Volume or Image must be specified"))
	}
	return errs
}

func (r *DataDestination) validateUpdate(old *DataDestination) (errs *apis.FieldError) {
	// TODO: Implement Volumes
	if r.Volume != nil {
		errs = errs.Also(apis.ErrInvalidValue("Volume support is not implemented yet", "Volume"))
	}
	if old.Image != r.Image {
		errs = errs.Also(apis.ErrInvalidValue("Image field cannot be changed once set", "Image"))
	}

	if old.ImagePushSecret != r.ImagePushSecret {
		errs = errs.Also(apis.ErrInvalidValue("ImagePushSecret field cannot be changed once set", "ImagePushSecret"))
	}
	return errs
}

func (r *ResourceSpec) validateCreateWithTuning(tuning *TuningSpec) (errs *apis.FieldError) {
	if *r.Count > 1 {
		errs = errs.Also(apis.ErrInvalidValue("Tuning does not currently support multinode configurations. Please set the node count to 1. Future support with DeepSpeed will allow this.", "count"))
	}
	return errs
}

func (r *ResourceSpec) validateCreateWithInference(inference *InferenceSpec) (errs *apis.FieldError) {
	var presetName string
	if inference.Preset != nil {
		presetName = strings.ToLower(string(inference.Preset.Name))
	}
	instanceType := string(r.InstanceType)

	// Check if instancetype exists in our SKUs map
	if skuConfig, exists := SupportedGPUConfigs[instanceType]; exists {
		if presetName != "" {
			model := plugin.KaitoModelRegister.MustGet(presetName) // InferenceSpec has been validated so the name is valid.
			// Validate GPU count for given SKU
			machineCount := *r.Count
			totalNumGPUs := machineCount * skuConfig.GPUCount
			totalGPUMem := machineCount * skuConfig.GPUMem * skuConfig.GPUCount

			modelGPUCount := resource.MustParse(model.GetInferenceParameters().GPUCountRequirement)
			modelPerGPUMemory := resource.MustParse(model.GetInferenceParameters().PerGPUMemoryRequirement)
			modelTotalGPUMemory := resource.MustParse(model.GetInferenceParameters().TotalGPUMemoryRequirement)

			// Separate the checks for specific error messages
			if int64(totalNumGPUs) < modelGPUCount.Value() {
				errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Insufficient number of GPUs: Instance type %s provides %d, but preset %s requires at least %d", instanceType, totalNumGPUs, presetName, modelGPUCount.Value()), "instanceType"))
			}
			skuPerGPUMemory := skuConfig.GPUMem / skuConfig.GPUCount
			if int64(skuPerGPUMemory) < modelPerGPUMemory.ScaledValue(resource.Giga) {
				errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Insufficient per GPU memory: Instance type %s provides %d per GPU, but preset %s requires at least %d per GPU", instanceType, skuPerGPUMemory, presetName, modelPerGPUMemory.ScaledValue(resource.Giga)), "instanceType"))
			}
			if int64(totalGPUMem) < modelTotalGPUMemory.ScaledValue(resource.Giga) {
				errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Insufficient total GPU memory: Instance type %s has a total of %d, but preset %s requires at least %d", instanceType, totalGPUMem, presetName, modelTotalGPUMemory.ScaledValue(resource.Giga)), "instanceType"))
			}
		}
	} else {
		// Check for other instance types pattern matches
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

	if i.Preset != nil {
		presetName := string(i.Preset.Name)
		// Validate preset name
		if !isValidPreset(presetName) {
			errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("Unsupported inference preset name %s", presetName), "presetName"))
		}
		// Validate private preset has private image specified
		if plugin.KaitoModelRegister.MustGet(string(i.Preset.Name)).GetInferenceParameters().ImageAccessMode == string(ModelImageAccessModePrivate) &&
			i.Preset.PresetMeta.AccessMode != ModelImageAccessModePrivate {
			errs = errs.Also(apis.ErrGeneric("This preset only supports private AccessMode, AccessMode must be private to continue"))
		}
		// Additional validations for Preset
		if i.Preset.PresetMeta.AccessMode == ModelImageAccessModePrivate && i.Preset.PresetOptions.Image == "" {
			errs = errs.Also(apis.ErrGeneric("When AccessMode is private, an image must be provided in PresetOptions"))
		}
		// Note: we don't enforce private access mode to have image secrets, in case anonymous pulling is enabled
	}
	if len(i.Adapters) > MaxAdaptersNumber {
		errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Number of Adapters exceeds the maximum limit, maximum of %s allowed", strconv.Itoa(MaxAdaptersNumber))))
	}

	// check if adapter names are duplicate
	if len(i.Adapters) > 0 {
		nameMap := make(map[string]bool)
		errs = errs.Also(validateDuplicateName(i.Adapters, nameMap))
	}

	return errs
}

func (i *InferenceSpec) validateUpdate(old *InferenceSpec) (errs *apis.FieldError) {
	if !reflect.DeepEqual(i.Preset, old.Preset) {
		errs = errs.Also(apis.ErrGeneric("field is immutable", "preset"))
	}
	// inference.template can be changed, but cannot be set/unset.
	if (i.Template != nil && old.Template == nil) || (i.Template == nil && old.Template != nil) {
		errs = errs.Also(apis.ErrGeneric("field cannot be unset/set if it was set/unset", "template"))
	}

	// check if adapter names are duplicate
	for _, adapter := range i.Adapters {
		errs = errs.Also(adapter.validateCreateorUpdate())
	}

	// check if adapter names are duplicate

	if len(i.Adapters) > 0 {
		nameMap := make(map[string]bool)
		errs = errs.Also(validateDuplicateName(i.Adapters, nameMap))
	}
	return errs
}

func validateDuplicateName(adapters []AdapterSpec, nameMap map[string]bool) (errs *apis.FieldError) {
	for _, adapter := range adapters {
		if _, ok := nameMap[adapter.Source.Name]; ok {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Duplicate adapter source name found: %s", adapter.Source.Name)))
		} else {
			nameMap[adapter.Source.Name] = true
		}
	}
	return errs
}
