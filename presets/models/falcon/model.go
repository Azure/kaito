// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package falcon

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
		Name:     PresetFalcon7BModel,
		Instance: &falconA,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     PresetFalcon7BInstructModel,
		Instance: &falconB,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     PresetFalcon40BModel,
		Instance: &falconC,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     PresetFalcon40BInstructModel,
		Instance: &falconD,
	})
}

var (
	registryName = os.Getenv("PRESET_REGISTRY_NAME")

	PresetFalcon7BModel          = "falcon-7b"
	PresetFalcon40BModel         = "falcon-40b"
	PresetFalcon7BInstructModel  = PresetFalcon7BModel + "-instruct"
	PresetFalcon40BInstructModel = PresetFalcon40BModel + "-instruct"

	presetFalcon7bImage         = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetFalcon7BModel)
	presetFalcon7bInstructImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetFalcon7BInstructModel)

	presetFalcon40bImage         = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetFalcon40BModel)
	presetFalcon40bInstructImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetFalcon40BInstructModel)

	baseCommandPresetFalcon = "accelerate launch --use_deepspeed"
	falconRunParams         = map[string]string{}
)

var falconA falcon7b

type falcon7b struct{}

func (*falcon7b) GetImageInfo() *model.PresetImageInfo {
	return &model.PresetImageInfo{
		Image:            presetFalcon7bImage,
		ImagePullSecrets: inference.DefaultImagePullSecrets,
		ImageAccessMode:  inference.DefaultImageAccessMode,
	}
}

func (*falcon7b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "Falcon",
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "14Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            falconRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetFalcon,
	}

}
func (*falcon7b) SupportDistributedInference() bool {
	return false
}

var falconB falcon7bInst

type falcon7bInst struct{}

func (*falcon7bInst) GetImageInfo() *model.PresetImageInfo {
	return &model.PresetImageInfo{
		Image:            presetFalcon7bInstructImage,
		ImagePullSecrets: inference.DefaultImagePullSecrets,
		ImageAccessMode:  inference.DefaultImageAccessMode,
	}
}

func (*falcon7bInst) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "Falcon",
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "14Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            falconRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetFalcon,
	}

}
func (*falcon7bInst) SupportDistributedInference() bool {
	return false
}

var falconC falcon40b

type falcon40b struct{}

func (*falcon40b) GetImageInfo() *model.PresetImageInfo {
	return &model.PresetImageInfo{
		Image:            presetFalcon40bImage,
		ImagePullSecrets: inference.DefaultImagePullSecrets,
		ImageAccessMode:  inference.DefaultImageAccessMode,
	}
}

func (*falcon40b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "Falcon",
		DiskStorageRequirement:    "400",
		GPUCountRequirement:       "2",
		TotalGPUMemoryRequirement: "90Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            falconRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetFalcon,
	}

}
func (*falcon40b) SupportDistributedInference() bool {
	return false
}

var falconD falcon40bInst

type falcon40bInst struct{}

func (*falcon40bInst) GetImageInfo() *model.PresetImageInfo {
	return &model.PresetImageInfo{
		Image:            presetFalcon40bInstructImage,
		ImagePullSecrets: inference.DefaultImagePullSecrets,
		ImageAccessMode:  inference.DefaultImageAccessMode,
	}
}

func (*falcon40bInst) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "Falcon",
		DiskStorageRequirement:    "400",
		GPUCountRequirement:       "2",
		TotalGPUMemoryRequirement: "90Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            falconRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetFalcon,
	}
}

func (*falcon40bInst) SupportDistributedInference() bool {
	return false
}
