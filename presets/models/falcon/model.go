// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package falcon

import (
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
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
	PresetFalcon7BModel          = "falcon-7b"
	PresetFalcon40BModel         = "falcon-40b"
	PresetFalcon7BInstructModel  = PresetFalcon7BModel + "-instruct"
	PresetFalcon40BInstructModel = PresetFalcon40BModel + "-instruct"

	PresetFalconTagMap = map[string]string{
		"Falcon7B":          "0.0.1",
		"Falcon7BInstruct":  "0.0.1",
		"Falcon40B":         "0.0.1",
		"Falcon40BInstruct": "0.0.1",
	}

	baseCommandPresetFalcon = "accelerate launch --use_deepspeed"
	falconRunParams         = map[string]string{}

	/* TODO: Migrate to following for Falcon Image 0.0.2
		baseCommandPresetFalcon = "accelerate launch"

		falconRunParams = map[string]string{
			"torch_dtype": "bfloat16",
			"pipeline":    "text-generation",
		}
	*/
)

var falconA falcon7b

type falcon7b struct{}

func (*falcon7b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "Falcon",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "14Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            falconRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetFalcon,
		Tag:                       PresetFalconTagMap["Falcon7B"],
	}

}
func (*falcon7b) SupportDistributedInference() bool {
	return false
}

var falconB falcon7bInst

type falcon7bInst struct{}

func (*falcon7bInst) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "Falcon",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "14Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            falconRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetFalcon,
		Tag:                       PresetFalconTagMap["Falcon7BInstruct"],
	}

}
func (*falcon7bInst) SupportDistributedInference() bool {
	return false
}

var falconC falcon40b

type falcon40b struct{}

func (*falcon40b) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "Falcon",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "400",
		GPUCountRequirement:       "2",
		TotalGPUMemoryRequirement: "90Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            falconRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetFalcon,
		Tag:                       PresetFalconTagMap["Falcon40B"],
	}

}
func (*falcon40b) SupportDistributedInference() bool {
	return false
}

var falconD falcon40bInst

type falcon40bInst struct{}

func (*falcon40bInst) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		ModelFamilyName:           "Falcon",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "400",
		GPUCountRequirement:       "2",
		TotalGPUMemoryRequirement: "90Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Falcon using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            falconRunParams,
		DeploymentTimeout:         time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetFalcon,
		Tag:                       PresetFalconTagMap["Falcon40BInstruct"],
	}
}

func (*falcon40bInst) SupportDistributedInference() bool {
	return false
}
