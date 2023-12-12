// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package llama2chat

import (
	"time"

	"github.com/azure/kaito/pkg/inference"
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
	llamaChatInferenceFile = "inference-api.py"
)

var llama2chatA llama2Chat7b

type llama2Chat7b struct{}

func (*llama2Chat7b) GetInferenceParameters() *inference.PresetInferenceParam {
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
func (*llama2Chat7b) NeedStatefulSet() bool {
	return true
}
func (*llama2Chat7b) NeedHeadlessService() bool {
	return false
}

var llama2chatB llama2Chat13b

type llama2Chat13b struct{}

func (*llama2Chat13b) GetInferenceParameters() *inference.PresetInferenceParam {
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
func (*llama2Chat13b) NeedStatefulSet() bool {
	return true
}
func (*llama2Chat13b) NeedHeadlessService() bool {
	return true
}

var llama2chatC llama2Chat70b

type llama2Chat70b struct{}

func (*llama2Chat70b) GetInferenceParameters() *inference.PresetInferenceParam {
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
func (*llama2Chat70b) NeedStatefulSet() bool {
	return true
}
func (*llama2Chat70b) NeedHeadlessService() bool {
	return true
}
