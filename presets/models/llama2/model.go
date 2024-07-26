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
	baseCommandPresetLlama = "cd /workspace/llama/llama-2 && torchrun"
	llamaRunParams         = map[string]string{
		"max_seq_len":    "512",
		"max_batch_size": "8",
	}
)

var llama2A llama2Text7b

type llama2Text7b struct{}

func (*llama2Text7b) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "34G",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "14G",
		PerGPUMemoryRequirement:   "14G", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		ReadinessTimeout:          time.Duration(10) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 1,
		// Tag:  llama has private image access mode. The image tag is determined by the user.
	}

}
func (*llama2Text7b) GetTuningParameters() *model.PresetParam {
	return nil // Currently doesn't support fine-tuning
}
func (*llama2Text7b) SupportDistributedInference() bool {
	return false
}
func (*llama2Text7b) SupportTuning() bool {
	return false
}

var llama2B llama2Text13b

type llama2Text13b struct{}

func (*llama2Text13b) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "46G",
		GPUCountRequirement:       "2",
		TotalGPUMemoryRequirement: "30G",
		PerGPUMemoryRequirement:   "15G", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		ReadinessTimeout:          time.Duration(20) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 2,
		// Tag:  llama has private image access mode. The image tag is determined by the user.
	}
}
func (*llama2Text13b) GetTuningParameters() *model.PresetParam {
	return nil // Currently doesn't support fine-tuning
}
func (*llama2Text13b) SupportDistributedInference() bool {
	return true
}
func (*llama2Text13b) SupportTuning() bool {
	return false
}

var llama2C llama2Text70b

type llama2Text70b struct{}

func (*llama2Text70b) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "158G",
		GPUCountRequirement:       "8",
		TotalGPUMemoryRequirement: "152G",
		PerGPUMemoryRequirement:   "19G", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		ReadinessTimeout:          time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 8,
		// Tag:  llama has private image access mode. The image tag is determined by the user.
	}
}
func (*llama2Text70b) GetTuningParameters() *model.PresetParam {
	return nil // Currently doesn't support fine-tuning
}
func (*llama2Text70b) SupportDistributedInference() bool {
	return true
}
func (*llama2Text70b) SupportTuning() bool {
	return false
}
