// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package phi_3

import (
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/inference"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils/plugin"
)

func init() {
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     PresetPhi3Mini4kModel,
		Instance: &phi3MiniA,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     PresetPhi3Mini128kModel,
		Instance: &phi3MiniB,
	})
}

var (
	PresetPhi3Mini4kModel   = "phi3Mini4KInst"
	PresetPhi3Mini128kModel = "phi3Mini128KInst"

	PresetPhiTagMap = map[string]string{
		"Phi3Mini4kInstruct":   "0.0.1",
		"Phi3Mini128kInstruct": "0.0.1",
	}

	baseCommandPresetPhi = "accelerate launch"
	phiRunParams         = map[string]string{
		"torch_dtype":       "auto",
		"pipeline":          "text-generation",
		"trust_remote_code": "",
	}
)

var phi3MiniA phi3Mini4KInst

type phi3Mini4KInst struct{}

func (*phi3Mini4KInst) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Phi",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "9Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Phi using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            phiRunParams,
		ReadinessTimeout:          time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetPhi,
		Tag:                       PresetPhiTagMap["Phi3Mini4kInstruct"],
	}
}
func (*phi3Mini4KInst) GetTuningParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Phi",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16Gi",
		PerGPUMemoryRequirement:   "16Gi", // We run Phi using native vertical model parallel, no per GPU memory requirement.
		// TorchRunParams:            inference.DefaultAccelerateParams,
		// ModelRunParams:            phiRunParams,
		ReadinessTimeout: time.Duration(30) * time.Minute,
		BaseCommand:      baseCommandPresetPhi,
		Tag:              PresetPhiTagMap["Phi3Mini4kInstruct"],
	}
}
func (*phi3Mini4KInst) SupportDistributedInference() bool { return false }
func (*phi3Mini4KInst) SupportTuning() bool {
	return true
}

var phi3MiniB phi3Mini128KInst

type phi3Mini128KInst struct{}

func (*phi3Mini128KInst) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Phi",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "9Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Phi using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            phiRunParams,
		ReadinessTimeout:          time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetPhi,
		Tag:                       PresetPhiTagMap["Phi3Mini128kInstruct"],
	}
}
func (*phi3Mini128KInst) GetTuningParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Phi",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "16Gi",
		PerGPUMemoryRequirement:   "16Gi", // We run Phi using native vertical model parallel, no per GPU memory requirement.
		// TorchRunParams:            inference.DefaultAccelerateParams,
		// ModelRunParams:            phiRunParams,
		ReadinessTimeout: time.Duration(30) * time.Minute,
		BaseCommand:      baseCommandPresetPhi,
		Tag:              PresetPhiTagMap["Phi3Mini128kInstruct"],
	}
}
func (*phi3Mini128KInst) SupportDistributedInference() bool { return false }
func (*phi3Mini128KInst) SupportTuning() bool {
	return true
}
