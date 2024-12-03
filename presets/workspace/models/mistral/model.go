// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package mistral

import (
	"time"

	kaitov1alpha1 "github.com/kaito-project/kaito/api/v1alpha1"
	"github.com/kaito-project/kaito/pkg/model"
	"github.com/kaito-project/kaito/pkg/utils/plugin"
	"github.com/kaito-project/kaito/pkg/workspace/inference"
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
		"Mistral7B":         "0.0.8",
		"Mistral7BInstruct": "0.0.8",
	}

	baseCommandPresetMistralInference = "accelerate launch"
	baseCommandPresetMistralTuning    = "cd /workspace/tfs/ && python3 metrics_server.py & accelerate launch"
	mistralRunParams                  = map[string]string{
		"torch_dtype":   "bfloat16",
		"pipeline":      "text-generation",
		"chat_template": "/workspace/chat_templates/mistral-instruct.jinja",
	}
	mistralRunParamsVLLM = map[string]string{
		"dtype":         "float16",
		"chat-template": "/workspace/chat_templates/mistral-instruct.jinja",
	}
)

var mistralA mistral7b

type mistral7b struct{}

func (*mistral7b) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Mistral",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "100Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "14Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Mistral using native vertical model parallel, no per GPU memory requirement.
		RuntimeParam: model.RuntimeParam{
			Transformers: model.HuggingfaceTransformersParam{
				TorchRunParams:    inference.DefaultAccelerateParams,
				ModelRunParams:    mistralRunParams,
				BaseCommand:       baseCommandPresetMistralInference,
				InferenceMainFile: inference.DefautTransformersMainFile,
			},
			VLLM: model.VLLMParam{
				BaseCommand:    inference.DefaultVLLMCommand,
				ModelName:      "mistral-7b",
				ModelRunParams: mistralRunParamsVLLM,
			},
		},
		ReadinessTimeout: time.Duration(30) * time.Minute,
		Tag:              PresetMistralTagMap["Mistral7B"],
	}

}
func (*mistral7b) GetTuningParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Mistral",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "100Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16Gi",
		PerGPUMemoryRequirement:   "16Gi",
		RuntimeParam: model.RuntimeParam{
			Transformers: model.HuggingfaceTransformersParam{
				//TorchRunParams:            tuning.DefaultAccelerateParams,
				//ModelRunParams:            mistralRunParams,
				BaseCommand: baseCommandPresetMistralTuning,
			},
		},
		ReadinessTimeout: time.Duration(30) * time.Minute,
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
		DiskStorageRequirement:    "100Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run mistral using native vertical model parallel, no per GPU memory requirement.
		RuntimeParam: model.RuntimeParam{
			Transformers: model.HuggingfaceTransformersParam{
				TorchRunParams:    inference.DefaultAccelerateParams,
				ModelRunParams:    mistralRunParams,
				BaseCommand:       baseCommandPresetMistralInference,
				InferenceMainFile: inference.DefautTransformersMainFile,
			},
			VLLM: model.VLLMParam{
				BaseCommand:    inference.DefaultVLLMCommand,
				ModelName:      "mistral-7b-instruct",
				ModelRunParams: mistralRunParamsVLLM,
			},
		},
		ReadinessTimeout: time.Duration(30) * time.Minute,
		Tag:              PresetMistralTagMap["Mistral7BInstruct"],
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
