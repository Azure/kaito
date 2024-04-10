package tuning

import "fmt"
import "gopkg.in/yaml.v2"

type TrainingConfig struct {
	ModelConfig        map[string]interface{} `yaml:"ModelConfig"`
	TokenizerParams    map[string]interface{} `yaml:"TokenizerParams"`
	QuantizationConfig map[string]interface{} `yaml:"QuantizationConfig"`
	LoraConfig         map[string]interface{} `yaml:"LoraConfig"`
	TrainingArguments  map[string]interface{} `yaml:"TrainingArguments"`
	DatasetConfig      map[string]interface{} `yaml:"DatasetConfig"`
	DataCollator       map[string]interface{} `yaml:"DataCollator"`
}

// ParseTrainingConfig parses the YAML string into a nested map
func ParseTrainingConfig(trainingConfigStr string) (map[string]map[string]string, error) {
	var trainingConfigWrapper struct {
		TrainingConfig TrainingConfig `yaml:"training_config"`
	}

	err := yaml.Unmarshal([]byte(trainingConfigStr), &trainingConfigWrapper)
	if err != nil {
		return nil, err // handle parsing error
	}

	// Convert to map[string]map[string]string
	result := make(map[string]map[string]string)
	for section, params := range map[string]map[string]interface{}{
		"ModelConfig":        trainingConfigWrapper.TrainingConfig.ModelConfig,
		"TokenizerParams":    trainingConfigWrapper.TrainingConfig.TokenizerParams,
		"QuantizationConfig": trainingConfigWrapper.TrainingConfig.QuantizationConfig,
		"LoraConfig":         trainingConfigWrapper.TrainingConfig.LoraConfig,
		"TrainingArguments":  trainingConfigWrapper.TrainingConfig.TrainingArguments,
		"DatasetConfig":      trainingConfigWrapper.TrainingConfig.DatasetConfig,
		"DataCollator":       trainingConfigWrapper.TrainingConfig.DataCollator,
	} {
		sectionMap := make(map[string]string)
		for key, value := range params {
			// Assuming all values can be represented as strings
			sectionMap[key] = fmt.Sprint(value)
		}
		result[section] = sectionMap
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
