// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package mistral

import (
	"fmt"
	"os"
	"time"

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
	registryName = os.Getenv("PRESET_REGISTRY_NAME")

	PresetMistral7BModel         = "mistral-7b-v0.1"
	PresetMistral7BInstructModel = "mistral-7b-instruct-v0.1"

	presetMistral7bImage         = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetMistral7BModel)
	presetMistral7bInstructImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetMistral7BInstructModel)

	baseCommandPresetMistral = "accelerate launch --use_deepspeed"
	mistralInferenceFile     = "inference-api.py"
	mistralARunParams        = map[string]string{
		"torch_dtype": "float16",
		"pipeline":    "text-generation",
	}
	mistralBRunParams = map[string]string{
		"torch_dtype": "float16",
		"pipeline":    "conversational",
	}
)

var mistralA mistral7b

type mistral7b struct{}

func (*mistral7b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelName:                 "Mistral",
		Image:                     presetMistral7bImage,
		ImagePullSecrets:          inference.DefaultImagePullSecrets,
		AccessMode:                inference.DefaultAccessMode,
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Mistral using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            mistralARunParams,
		InferenceFile:             mistralInferenceFile,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetMistral,
		DefaultVolumeMountPath:    "/dev/shm",
	}

}
func (*mistral7b) SupportDistributedInference() bool {
	return false
}

var mistralB mistral7bInst

type mistral7bInst struct{}

func (*mistral7bInst) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelName:                 "Mistral",
		Image:                     presetMistral7bInstructImage,
		ImagePullSecrets:          inference.DefaultImagePullSecrets,
		AccessMode:                inference.DefaultAccessMode,
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            mistralBRunParams,
		InferenceFile:             mistralInferenceFile,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetMistral,
		DefaultVolumeMountPath:    "/dev/shm",
	}

}
func (*mistral7bInst) SupportDistributedInference() bool {
	return false
}
