package tuning

import (
	"fmt"
	"github.com/azure/kaito/pkg/config"
	"reflect"
)
import "gopkg.in/yaml.v2"

// ParseTrainingConfig parses the YAML string into a nested map
func ParseTrainingConfig(trainingConfigStr string) (map[string]map[string]string, error) {
	var trainingConfigWrapper struct {
		TrainingConfig config.TrainingConfig `yaml:"training_config"`
	}

	err := yaml.Unmarshal([]byte(trainingConfigStr), &trainingConfigWrapper)
	if err != nil {
		return nil, err // handle parsing error
	}

	// Convert to map[string]map[string]string
	result := make(map[string]map[string]string)

	trainingConfigVal := reflect.ValueOf(trainingConfigWrapper.TrainingConfig)
	for i := 0; i < trainingConfigVal.NumField(); i++ {
		field := trainingConfigVal.Field(i)
		if field.IsNil() {
			continue // Skip if the entire section is nil
		}

		sectionName := trainingConfigVal.Type().Field(i).Tag.Get("yaml")
		sectionMap := make(map[string]string)

		// Reflect over each field in the section
		for j := 0; j < field.Elem().NumField(); j++ {
			innerField := field.Elem().Field(j)
			key := field.Elem().Type().Field(j).Name

			// Skip nil pointer fields or decide based on zero value
			if innerField.Kind() == reflect.Ptr && innerField.IsNil() {
				continue // Skip nil pointer fields
			}

			// Convert non-nil value to string
			value := fmt.Sprint(innerField.Interface())
			sectionMap[key] = value
		}

		if len(sectionMap) > 0 {
			result[sectionName] = sectionMap
		}
	}
	return result, nil
}

func AddPrefixesToConfigMap(configMap map[string]map[string]string) (map[string]string, error) {
	prefixedConfigMap := make(map[string]string)
	for section, params := range configMap {
		prefix, err := GetCmdPrefixForSection(section)
		if err != nil {
			return nil, err
		}
		for param, value := range params {
			prefixedKey := fmt.Sprintf("%s_%s", prefix, param)
			prefixedConfigMap[prefixedKey] = value
		}
	}
	return prefixedConfigMap, nil
}

func GetCmdPrefixForSection(section string) (string, error) {
	prefixMap := map[string]string{
		"ModelConfig":        "MC",
		"QuantizationConfig": "QC",
		"LoraConfig":         "ELC",
		"TrainingArguments":  "TA",
		"DataCollator":       "EDC",
		"DatasetConfig":      "DC",
		"TokenizerParams":    "TP",
	}

	if prefix, ok := prefixMap[section]; ok {
		return prefix, nil
	}
	return "", fmt.Errorf("prefix for section '%s' not found", section)
}
