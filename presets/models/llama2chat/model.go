// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package llama2chat

import (
	"time"

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

func (*llama2Chat7b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "LLaMa2",
		Image:                     "",
		ImagePullSecrets:          inference.DefaultImagePullSecrets,
		ImageAccessMode:           inference.DefaultImageAccessMode,
		DiskStorageRequirement:    "34Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16Gi",
		PerGPUMemoryRequirement:   "14Gi", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		DeploymentTimeout:         time.Duration(10) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 1,
	}

}
func (*llama2Chat7b) SupportDistributedInference() bool {
	return false
}

var llama2chatB llama2Chat13b

type llama2Chat13b struct{}

func (*llama2Chat13b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "LLaMa2",
		Image:                     "",
		ImagePullSecrets:          inference.DefaultImagePullSecrets,
		ImageAccessMode:           inference.DefaultImageAccessMode,
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
	}
}
func (*llama2Chat13b) SupportDistributedInference() bool {
	return true
}

var llama2chatC llama2Chat70b

type llama2Chat70b struct{}

func (*llama2Chat70b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "LLaMa2",
		Image:                     "",
		ImagePullSecrets:          inference.DefaultImagePullSecrets,
		ImageAccessMode:           inference.DefaultImageAccessMode,
		DiskStorageRequirement:    "158Gi",
		GPUCountRequirement:       "8",
		TotalGPUMemoryRequirement: "192Gi",
		PerGPUMemoryRequirement:   "19Gi", // We run llama2 using tensor parallelism, the memory of each GPU needs to be bigger than the tensor shard size.
		TorchRunParams:            inference.DefaultTorchRunParams,
		TorchRunRdzvParams:        inference.DefaultTorchRunRdzvParams,
		ModelRunParams:            llamaRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetLlama,
		WorldSize:                 8,
	}
}
func (*llama2Chat70b) SupportDistributedInference() bool {
	return true
}
