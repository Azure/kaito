// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

import (
	"context"
	"fmt"
	"reflect"

	"github.com/azure/kaito/pkg/k8sclient"
	"github.com/azure/kaito/pkg/utils"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	TrainingConfig TrainingConfig `yaml:"training_config"`
}

type TrainingConfig struct {
	ModelConfig        map[string]runtime.RawExtension `yaml:"ModelConfig"`
	TokenizerParams    map[string]runtime.RawExtension `yaml:"TokenizerParams"`
	QuantizationConfig map[string]runtime.RawExtension `yaml:"QuantizationConfig"`
	LoraConfig         map[string]runtime.RawExtension `yaml:"LoraConfig"`
	TrainingArguments  map[string]runtime.RawExtension `yaml:"TrainingArguments"`
	DatasetConfig      map[string]runtime.RawExtension `yaml:"DatasetConfig"`
	DataCollator       map[string]runtime.RawExtension `yaml:"DataCollator"`
}

func validateNilOrBool(value interface{}) error {
	if value == nil {
		return nil // nil is acceptable
	}
	if _, ok := value.(bool); ok {
		return nil // Correct type
	}
	return fmt.Errorf("value must be either nil or a boolean, got type %T", value)
}

// UnmarshalYAML custom method
func (t *TrainingConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	// This function converts a map[string]interface{} to a map[string]runtime.RawExtension.
	// It does this by setting the raw marshalled data of the unmarshalled YAML to
	// be the raw data of the runtime.RawExtension object.
	handleRawExtension := func(raw map[string]interface{}, field string, target map[string]runtime.RawExtension) error {
		if value, found := raw[field]; found {
			delete(raw, field)
			var ext runtime.RawExtension
			data, err := yaml.Marshal(value)
			if err != nil {
				return err
			}
			ext.Raw = data
			if target == nil {
				target = make(map[string]runtime.RawExtension)
			}
			target[field] = ext
		}
		return nil
	}

	fields := []struct {
		name   string
		target map[string]runtime.RawExtension
	}{
		{"ModelConfig", t.ModelConfig},
		{"TokenizerParams", t.TokenizerParams},
		{"QuantizationConfig", t.QuantizationConfig},
		{"LoraConfig", t.LoraConfig},
		{"TrainingArguments", t.TrainingArguments},
		{"DatasetConfig", t.DatasetConfig},
		{"DataCollator", t.DataCollator},
	}

	for _, field := range fields {
		if err := handleRawExtension(raw, field.name, field.target); err != nil {
			return err
		}
	}

	return nil
}

func validateMethodViaConfigMap(cm *corev1.ConfigMap, methodLowerCase string) *apis.FieldError {
	trainingConfigYAML, ok := cm.Data["training_config.yaml"]
	if !ok {
		return apis.ErrGeneric(fmt.Sprintf("ConfigMap '%s' does not contain 'training_config.yaml' in namespace '%s'", cm.Name, cm.Namespace), "config")
	}

	var config Config
	if err := yaml.Unmarshal([]byte(trainingConfigYAML), &config); err != nil {
		return apis.ErrGeneric(fmt.Sprintf("Failed to parse 'training_config.yaml' in ConfigMap '%s' in namespace '%s': %v", cm.Name, cm.Namespace, err), "config")
	}

	// Validate QuantizationConfig if it exists
	quantConfig := config.TrainingConfig.QuantizationConfig
	if quantConfig != nil {
		quantConfigRaw, quantConfigExists := quantConfig["QuantizationConfig"]
		if quantConfigExists {
			// Dynamic field search for quantization settings within ModelConfig
			loadIn4bit, _, err := utils.SearchRawExtension(quantConfigRaw, "load_in_4bit")
			if err != nil {
				return apis.ErrInvalidValue(err.Error(), "load_in_4bit")
			}
			loadIn8bit, _, err := utils.SearchRawExtension(quantConfigRaw, "load_in_8bit")
			if err != nil {
				return apis.ErrInvalidValue(err.Error(), "load_in_8bit")
			}

			// Validate both loadIn4bit and loadIn8bit
			if err := validateNilOrBool(loadIn4bit); err != nil {
				return apis.ErrInvalidValue(err.Error(), "load_in_4bit")
			}
			if err := validateNilOrBool(loadIn8bit); err != nil {
				return apis.ErrInvalidValue(err.Error(), "load_in_8bit")
			}

			loadIn4bitBool, _ := loadIn4bit.(bool)
			loadIn8bitBool, _ := loadIn8bit.(bool)

			// Validation Logic
			if loadIn4bitBool && loadIn8bitBool {
				return apis.ErrGeneric(fmt.Sprintf("Cannot set both 'load_in_4bit' and 'load_in_8bit' to true in ConfigMap '%s'", cm.Name), "QuantizationConfig")
			}
			if methodLowerCase == string(TuningMethodLora) {
				if loadIn4bitBool || loadIn8bitBool {
					return apis.ErrGeneric(fmt.Sprintf("For method 'lora', 'load_in_4bit' or 'load_in_8bit' in ConfigMap '%s' must not be true", cm.Name), "QuantizationConfig")
				}
			} else if methodLowerCase == string(TuningMethodQLora) {
				if !loadIn4bitBool && !loadIn8bitBool {
					return apis.ErrMissingField(fmt.Sprintf("For method 'qlora', either 'load_in_4bit' or 'load_in_8bit' must be true in ConfigMap '%s'", cm.Name), "QuantizationConfig")
				}
			}
		}
	} else if methodLowerCase == string(TuningMethodQLora) {
		return apis.ErrMissingField(fmt.Sprintf("For method 'qlora', either 'load_in_4bit' or 'load_in_8bit' must be true in ConfigMap '%s'", cm.Name), "QuantizationConfig")
	}
	return nil
}

// getStructInstances dynamically generates instances of all sections in any config struct.
func getStructInstances(s any) map[string]any {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem() // Dereference pointer to get the struct type
	}
	instances := make(map[string]any, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag != "" {
			// Create a new instance of the type pointed to by the field
			instance := reflect.MakeMap(field.Type).Interface()
			instances[yamlTag] = instance
		}
	}

	return instances
}

func validateConfigMapSchema(cm *corev1.ConfigMap) *apis.FieldError {
	trainingConfigData, ok := cm.Data["training_config.yaml"]
	if !ok {
		return apis.ErrMissingField("training_config.yaml in ConfigMap")
	}

	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal([]byte(trainingConfigData), &rawConfig); err != nil {
		return apis.ErrInvalidValue(err.Error(), "training_config.yaml")
	}

	// Extract the actual training configuration map
	trainingConfigMap, ok := rawConfig["training_config"].(map[interface{}]interface{})
	if !ok {
		return apis.ErrInvalidValue("Expected 'training_config' key to contain a map", "training_config.yaml")
	}

	sectionStructs := getStructInstances(TrainingConfig{})
	recognizedSections := make([]string, 0, len(sectionStructs))
	for section := range sectionStructs {
		recognizedSections = append(recognizedSections, section)
	}

	// Check if valid sections
	for section := range trainingConfigMap {
		sectionStr := section.(string)
		if !utils.Contains(recognizedSections, sectionStr) {
			return apis.ErrInvalidValue(fmt.Sprintf("Unrecognized section: %s", section), "training_config.yaml")
		}
	}
	return nil
}

func (r *TuningSpec) validateConfigMap(ctx context.Context, namespace string, methodLowerCase string, configMapName string) (errs *apis.FieldError) {
	var cm corev1.ConfigMap
	if k8sclient.Client == nil {
		errs = errs.Also(apis.ErrGeneric("Failed to obtain client from context.Context"))
		return errs
	}
	err := k8sclient.Client.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: namespace}, &cm)
	if err != nil {
		if errors.IsNotFound(err) {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("ConfigMap '%s' specified in 'config' not found in namespace '%s'", r.ConfigTemplate, namespace), "config"))
		} else {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Failed to get ConfigMap '%s' in namespace '%s': %v", r.ConfigTemplate, namespace, err), "config"))
		}
	} else {
		if err := validateConfigMapSchema(&cm); err != nil {
			errs = errs.Also(err)
		}
		if err := validateMethodViaConfigMap(&cm, methodLowerCase); err != nil {
			errs = errs.Also(err)
		}
	}
	return errs
}
