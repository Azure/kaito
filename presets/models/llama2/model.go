// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package llama2

import (
	"time"

	"github.com/azure/kaito/pkg/inference"
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
	llamaChatInferenceFile = "inference-api.py"
)

var llama2A llama2Text7b

type llama2Text7b struct{}

func (*llama2Text7b) GetInferenceParameters() *inference.PresetInferenceParam {
	return &inference.PresetInferenceParam{
		ModelName:              "LLaMa2",
		Image:                  "",
		ImagePullSecrets:       inference.DefaultImagePullSecrets,
		AccessMode:             inference.DefaultAccessMode,
		DiskStorageRequirement: "34Gi",
		GPURequirement:         "1",
		GPUMemoryRequirement:   "16Gi",
		TorchRunParams:         inference.DefaultTorchRunParams,
		TorchRunRdzvParams:     inference.DefaultTorchRunRdzvParams,
		ModelRunParams:         llamaRunParams,
		InferenceFile:          llamaChatInferenceFile,
		DeploymentTimeout:      time.Duration(10) * time.Minute,
		BaseCommand:            baseCommandPresetLlama,
		WorldSize:              1,
		DefaultVolumeMountPath: "/dev/shm",
	}

}
func (*llama2Text7b) SupportDistributedInference() bool {
	return false
}

var llama2B llama2Text13b

type llama2Text13b struct{}

func (*llama2Text13b) GetInferenceParameters() *inference.PresetInferenceParam {
	return &inference.PresetInferenceParam{
		ModelName:              "LLaMa2",
		Image:                  "",
		ImagePullSecrets:       inference.DefaultImagePullSecrets,
		AccessMode:             inference.DefaultAccessMode,
		DiskStorageRequirement: "46Gi",
		GPURequirement:         "2",
		GPUMemoryRequirement:   "16Gi",
		TorchRunParams:         inference.DefaultTorchRunParams,
		TorchRunRdzvParams:     inference.DefaultTorchRunRdzvParams,
		ModelRunParams:         llamaRunParams,
		InferenceFile:          llamaChatInferenceFile,
		DeploymentTimeout:      time.Duration(20) * time.Minute,
		BaseCommand:            baseCommandPresetLlama,
		WorldSize:              2,
		DefaultVolumeMountPath: "/dev/shm",
	}
}
func (*llama2Text13b) SupportDistributedInference() bool {
	return true
}

var llama2C llama2Text70b

type llama2Text70b struct{}

func (*llama2Text70b) GetInferenceParameters() *inference.PresetInferenceParam {
	return &inference.PresetInferenceParam{
		ModelName:              "LLaMa2",
		Image:                  "",
		ImagePullSecrets:       inference.DefaultImagePullSecrets,
		AccessMode:             inference.DefaultAccessMode,
		DiskStorageRequirement: "158Gi",
		GPURequirement:         "8",
		GPUMemoryRequirement:   "19Gi",
		TorchRunParams:         inference.DefaultTorchRunParams,
		TorchRunRdzvParams:     inference.DefaultTorchRunRdzvParams,
		ModelRunParams:         llamaRunParams,
		InferenceFile:          llamaChatInferenceFile,
		DeploymentTimeout:      time.Duration(30) * time.Minute,
		BaseCommand:            baseCommandPresetLlama,
		WorldSize:              8,
		DefaultVolumeMountPath: "/dev/shm",
	}
}
func (*llama2Text70b) SupportDistributedInference() bool {
	return true
}
