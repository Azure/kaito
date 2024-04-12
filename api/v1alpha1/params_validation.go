package v1alpha1

import (
	"context"
	"fmt"
	"github.com/azure/kaito/pkg/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"knative.dev/pkg/apis"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	DefaultReleaseNamespaceEnvVar = "RELEASE_NAMESPACE"
)

func getReleaseNamespace() (string, error) {
	// Path to the namespace file inside a Kubernetes pod
	namespaceFilePath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	// Attempt to read the namespace from the file
	if content, err := ioutil.ReadFile(namespaceFilePath); err == nil {
		return string(content), nil
	}

	// Fallback: Read the namespace from an environment variable
	if namespace, exists := os.LookupEnv(DefaultReleaseNamespaceEnvVar); exists {
		return namespace, nil
	}
	return "", fmt.Errorf("failed to determine release namespace from file %s and env var %s", namespaceFilePath, DefaultReleaseNamespaceEnvVar)
}

func validateMethodViaConfigMap(cm *corev1.ConfigMap, methodLowerCase string) *apis.FieldError {
	trainingConfigYAML, ok := cm.Data["training_config.yaml"]
	if !ok {
		return apis.ErrGeneric(fmt.Sprintf("ConfigMap '%s' does not contain 'training_config.yaml' in namespace '%s'", cm.Name, cm.Namespace), "config")
	}

	var trainingConfig config.TrainingConfig
	if err := yaml.Unmarshal([]byte(trainingConfigYAML), &trainingConfig); err != nil {
		return apis.ErrGeneric(fmt.Sprintf("Failed to parse 'training_config.yaml' in ConfigMap '%s' in namespace '%s': %v", cm.Name, cm.Namespace, err), "config")
	}

	quantConfig, quantConfigExists := trainingConfig.QuantizationConfig, trainingConfig.QuantizationConfig != nil
	if quantConfigExists {
		loadIn4bit, ok4bit := *quantConfig.LoadIn4bit, quantConfig.LoadIn4bit != nil
		loadIn8bit, ok8bit := *quantConfig.LoadIn8bit, quantConfig.LoadIn8bit != nil
		if ok4bit && ok8bit && loadIn4bit && loadIn8bit {
			return apis.ErrGeneric(fmt.Sprintf("Cannot set both 'load_in_4bit' and 'load_in_8bit' to true in ConfigMap '%s'", cm.Name), "QuantizationConfig")
		}
		if methodLowerCase == string(TuningMethodLora) {
			if (ok4bit && loadIn4bit) || (ok8bit && loadIn8bit) {
				return apis.ErrGeneric(fmt.Sprintf("For method 'lora', 'load_in_4bit' or 'load_in_8bit' in ConfigMap '%s' must not be true", cm.Name), "QuantizationConfig")
			}
		} else if methodLowerCase == string(TuningMethodQLora) {
			if !(ok4bit && loadIn4bit) && !(ok8bit && loadIn8bit) {
				return apis.ErrMissingField(fmt.Sprintf("For method 'qlora', either 'load_in_4bit' or 'load_in_8bit' must be true in ConfigMap '%s'", cm.Name), "QuantizationConfig")
			}
		}
	} else if methodLowerCase == string(TuningMethodQLora) {
		return apis.ErrMissingField(fmt.Sprintf("For method 'qlora', either 'load_in_4bit' or 'load_in_8bit' must be true in ConfigMap '%s'", cm.Name), "QuantizationConfig")
	}
	return nil
}

// getTagsFromStruct returns a slice of yaml tag names for fields in the TrainingConfig struct.
func getTagsFromStruct(s any, tagKey string) []string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem() // Dereference pointer to get the struct type
	}
	tags := make([]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagKey)
		if tag != "" {
			tagName := strings.Split(tag, ",")[0] // Remove any tag options (like omitempty)
			tags = append(tags, tagName)
		}
	}

	return tags
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
		if yamlTag != "" && field.Type.Kind() == reflect.Ptr {
			// Create a new instance of the type pointed to by the field
			instance := reflect.New(field.Type.Elem()).Interface()
			instances[yamlTag] = instance
		}
	}

	return instances
}

// containsString checks if a string is present in a slice.
func containsString(key string, slice []string) bool {
	for _, item := range slice {
		if item == key {
			return true
		}
	}
	return false
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

	sectionStructs := getStructInstances(config.TrainingConfig{})
	recognizedSections := make([]string, 0, len(sectionStructs))
	for section := range sectionStructs {
		recognizedSections = append(recognizedSections, section)
	}

	for section, contents := range trainingConfigMap {
		sectionStr := section.(string)
		// Check if section is recognized
		if !containsString(sectionStr, recognizedSections) {
			return apis.ErrInvalidValue(fmt.Sprintf("Unrecognized section: %s", section), "training_config.yaml")
		}

		// Assuming contents is a map, check each field in the section
		fieldMap, ok := contents.(map[interface{}]interface{})
		if !ok {
			continue // or handle the error
		}

		structInstance, _ := sectionStructs[sectionStr]
		fieldTags := getTagsFromStruct(structInstance, "json")
		for key := range fieldMap {
			if !containsString(fmt.Sprint(key), fieldTags) {
				return apis.ErrInvalidValue(fmt.Sprintf("Unrecognized field: %s in section %s", key, section), "training_config.yaml")
			}

			// TODO? Here could also attempt to validate the type of fieldMap[key]
		}
	}
	return nil
}

func (r *TuningSpec) validateConfigMap(ctx context.Context, methodLowerCase string) (errs *apis.FieldError) {
	namespace, err := getReleaseNamespace()
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
