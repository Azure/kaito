// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package mistral

import (
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/inference"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils/plugin"
)

func init() {
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     PresetMistral7BModel,
		Instance: &mistralA,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     PresetMistral7BInstructModel,
		Instance: &mistralB,
	})
}

var (
	PresetMistral7BModel         = "mistral-7b"
	PresetMistral7BInstructModel = PresetMistral7BModel + "-instruct"

	PresetMistralTagMap = map[string]string{
		"Mistral7B":         "0.0.6",
		"Mistral7BInstruct": "0.0.6",
	}

	baseCommandPresetMistral = "accelerate launch"
	mistralRunParams         = map[string]string{
		"torch_dtype": "bfloat16",
		"pipeline":    "text-generation",
	}
)

var mistralA mistral7b

type mistral7b struct{}

func (*mistral7b) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Mistral",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "100G",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "14G",
		PerGPUMemoryRequirement:   "0G", // We run Mistral using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            mistralRunParams,
		ReadinessTimeout:          time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetMistral,
		Tag:                       PresetMistralTagMap["Mistral7B"],
	}

}
func (*mistral7b) GetTuningParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Mistral",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "100G",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16G",
		PerGPUMemoryRequirement:   "16G",
		//TorchRunParams:            tuning.DefaultAccelerateParams,
		//ModelRunParams:            mistralRunParams,
		ReadinessTimeout: time.Duration(30) * time.Minute,
		BaseCommand:      baseCommandPresetMistral,
		Tag:              PresetMistralTagMap["Mistral7B"],
	}
}

func (*mistral7b) SupportDistributedInference() bool {
	return false
}
func (*mistral7b) SupportTuning() bool {
	return true
}

var mistralB mistral7bInst

type mistral7bInst struct{}

func (*mistral7bInst) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Mistral",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "100G",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16G",
		PerGPUMemoryRequirement:   "0G", // We run mistral using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            mistralRunParams,
		ReadinessTimeout:          time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetMistral,
		Tag:                       PresetMistralTagMap["Mistral7BInstruct"],
	}

}
func (*mistral7bInst) GetTuningParameters() *model.PresetParam {
	return nil // It is not recommended/ideal to further fine-tune instruct models - Already been fine-tuned
}
func (*mistral7bInst) SupportDistributedInference() bool {
	return false
}
func (*mistral7bInst) SupportTuning() bool {
	return false
}
