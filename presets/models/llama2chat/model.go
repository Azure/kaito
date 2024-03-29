// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package llama2chat

import (
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/inference"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils/plugin"
)

func init() {
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "llama-2-7b-chat",
		Instance: &llama2chatA,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "llama-2-13b-chat",
		Instance: &llama2chatB,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "llama-2-70b-chat",
		Instance: &llama2chatC,
	})
}

var (
	baseCommandPresetLlama = "cd /workspace/llama/llama-2 && torchrun"
	llamaRunParams         = map[string]string{
		"max_seq_len":    "512",
		"max_batch_size": "8",
	}
)

var llama2chatA llama2Chat7b

type llama2Chat7b struct{}

func (*llama2Chat7b) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "34Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16Gi",
		PerGPUMemoryRequirement:   "14Gi", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		ReadinessTimeout:          time.Duration(10) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 1,
		// Tag:  llama has private image access mode. The image tag is determined by the user.
	}
}
func (*llama2Chat7b) GetTuningParameters() *model.PresetParam {
	return nil // Currently doesn't support fine-tuning
}
func (*llama2Chat7b) SupportDistributedInference() bool {
	return false
}
func (*llama2Chat7b) SupportTuning() bool {
	return false
}

var llama2chatB llama2Chat13b

type llama2Chat13b struct{}

func (*llama2Chat13b) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "46Gi",
		GPUCountRequirement:       "2",
		TotalGPUMemoryRequirement: "30Gi",
		PerGPUMemoryRequirement:   "15Gi", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		ReadinessTimeout:          time.Duration(20) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 2,
		// Tag:  llama has private image access mode. The image tag is determined by the user.
	}
}
func (*llama2Chat13b) GetTuningParameters() *model.PresetParam {
	return nil // Currently doesn't support fine-tuning
}
func (*llama2Chat13b) SupportDistributedInference() bool {
	return true
}
func (*llama2Chat13b) SupportTuning() bool {
	return false
}

var llama2chatC llama2Chat70b

type llama2Chat70b struct{}

func (*llama2Chat70b) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "LLaMa2",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePrivate),
		DiskStorageRequirement:    "158Gi",
		GPUCountRequirement:       "8",
		TotalGPUMemoryRequirement: "192Gi",
		PerGPUMemoryRequirement:   "19Gi", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		ReadinessTimeout:          time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 8,
		// Tag:  llama has private image access mode. The image tag is determined by the user.
	}
}
func (*llama2Chat70b) GetTuningParameters() *model.PresetParam {
	return nil // Currently doesn't support fine-tuning
}
func (*llama2Chat70b) SupportDistributedInference() bool {
	return true
}
func (*llama2Chat70b) SupportTuning() bool {
	return false
}
