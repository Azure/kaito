package v1alpha1

import (
	"context"
	"fmt"
	"github.com/azure/kaito/pkg/utils"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"knative.dev/pkg/apis"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TrainingConfig struct {
	ModelConfig        map[string]interface{} `yaml:"ModelConfig"`
	TokenizerParams    map[string]interface{} `yaml:"TokenizerParams"`
	QuantizationConfig map[string]interface{} `yaml:"QuantizationConfig"`
	LoraConfig         map[string]interface{} `yaml:"LoraConfig"`
	TrainingArguments  map[string]interface{} `yaml:"TrainingArguments"`
	DatasetConfig      map[string]interface{} `yaml:"DatasetConfig"`
	DataCollator       map[string]interface{} `yaml:"DataCollator"`
}

func validateMethodViaConfigMap(cm *corev1.ConfigMap, methodLowerCase string) *apis.FieldError {
	trainingConfigYAML, ok := cm.Data["training_config.yaml"]
	if !ok {
		return apis.ErrGeneric(fmt.Sprintf("ConfigMap '%s' does not contain 'training_config.yaml' in namespace '%s'", cm.Name, cm.Namespace), "config")
	}

	var trainingConfig TrainingConfig
	if err := yaml.Unmarshal([]byte(trainingConfigYAML), &trainingConfig); err != nil {
		return apis.ErrGeneric(fmt.Sprintf("Failed to parse 'training_config.yaml' in ConfigMap '%s' in namespace '%s': %v", cm.Name, cm.Namespace, err), "config")
	}

	// Validate QuantizationConfig if it exists
	quantConfig := trainingConfig.QuantizationConfig
	if quantConfig != nil {
		// Dynamic field search for quantization settings within ModelConfig
		loadIn4bit, _ := utils.SearchMap(quantConfig, "load_in_4bit")
		loadIn8bit, _ := utils.SearchMap(quantConfig, "load_in_8bit")

		loadIn4bitBool, ok4bitBool := loadIn4bit.(bool)
		loadIn8bitBool, ok8bitBool := loadIn8bit.(bool)

		if ok4bitBool && ok8bitBool && loadIn4bitBool && loadIn8bitBool {
			return apis.ErrGeneric(fmt.Sprintf("Cannot set both 'load_in_4bit' and 'load_in_8bit' to true in ConfigMap '%s'", cm.Name), "QuantizationConfig")
		}
		if methodLowerCase == string(TuningMethodLora) {
			if (ok4bitBool && loadIn4bitBool) || (ok8bitBool && loadIn8bitBool) {
				return apis.ErrGeneric(fmt.Sprintf("For method 'lora', 'load_in_4bit' or 'load_in_8bit' in ConfigMap '%s' must not be true", cm.Name), "QuantizationConfig")
			}
		} else if methodLowerCase == string(TuningMethodQLora) {
			if !(ok4bitBool && loadIn4bitBool) && !(ok8bitBool && loadIn8bitBool) {
				return apis.ErrMissingField(fmt.Sprintf("For method 'qlora', either 'load_in_4bit' or 'load_in_8bit' must be true in ConfigMap '%s'", cm.Name), "QuantizationConfig")
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
	for section, _ := range trainingConfigMap {
		sectionStr := section.(string)
		if !utils.Contains(recognizedSections, sectionStr) {
			return apis.ErrInvalidValue(fmt.Sprintf("Unrecognized section: %s", section), "training_config.yaml")
		}
	}
	return nil
}

func (r *TuningSpec) validateConfigMap(ctx context.Context, methodLowerCase string) (errs *apis.FieldError) {
	namespace, err := utils.GetReleaseNamespace()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to determine release namespace: %v", err)
		errs = errs.Also(apis.ErrGeneric(errMsg, "namespace"))
	}

	var cm corev1.ConfigMap
	cl := getTestClient(ctx)
	if cl == nil {
		errMsg := fmt.Sprintf("Failed to obtain client from context.Context")
		errs = errs.Also(apis.ErrGeneric(errMsg))
		return errs
	}
	err = cl.Get(ctx, client.ObjectKey{Name: r.Config, Namespace: namespace}, &cm)
	if err != nil {
		if errors.IsNotFound(err) {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("ConfigMap '%s' specified in 'config' not found in namespace '%s'", r.Config, namespace), "config"))
		} else {
			errs = errs.Also(apis.ErrGeneric(fmt.Sprintf("Failed to get ConfigMap '%s' in namespace '%s': %v", r.Config, namespace, err), "config"))
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
