// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package llama2

import (
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/inference"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils/plugin"
)

func init() {
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "llama-2-7b",
		Instance: &llama2A,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "llama-2-13b",
		Instance: &llama2B,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "llama-2-70b",
		Instance: &llama2C,
	})
}

var (
	PresetLlamaTagMap = map[string]string{
		"llama-2-7b":  "0.0.3",
		"llama-2-13b": "0.0.3",
		"llama-2-70b": "0.0.3",
	}

	baseCommandPresetLlama = "cd /workspace/llama/llama-2 && torchrun"
	llamaRunParams         = map[string]string{
		"max_seq_len":    "512",
		"max_batch_size": "8",
	}
)

var llama2A llama2Text7b

type llama2Text7b struct{}

func (*llama2Text7b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "34Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "14Gi",
		PerGPUMemoryRequirement:   "14Gi", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		DeploymentTimeout:         time.Duration(10) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 1,
		Tag:                       PresetLlamaTagMap["llama-2-7b"],
	}

}
func (*llama2Text7b) SupportDistributedInference() bool {
	return false
}

var llama2B llama2Text13b

type llama2Text13b struct{}

func (*llama2Text13b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "46Gi",
		GPUCountRequirement:       "2",
		TotalGPUMemoryRequirement: "30Gi",
		PerGPUMemoryRequirement:   "15Gi", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		DeploymentTimeout:         time.Duration(20) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 2,
		Tag:                       PresetLlamaTagMap["llama-2-13b"],
	}
}
func (*llama2Text13b) SupportDistributedInference() bool {
	return true
}

var llama2C llama2Text70b

type llama2Text70b struct{}

func (*llama2Text70b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "158Gi",
		GPUCountRequirement:       "8",
		TotalGPUMemoryRequirement: "152Gi",
		PerGPUMemoryRequirement:   "19Gi", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 8,
		Tag:                       PresetLlamaTagMap["llama-2-70b"],
	}
}
func (*llama2Text70b) SupportDistributedInference() bool {
	return true
}
