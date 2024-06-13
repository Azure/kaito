// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package phi_2

import (
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/inference"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils/plugin"
)

func init() {
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     PresetPhi2Model,
		Instance: &phiA,
	})
}

var (
	PresetPhi2Model = "phi-2"

	PresetPhiTagMap = map[string]string{
		"Phi2": "0.0.3",
	}

	baseCommandPresetPhi = "accelerate launch"
	phiRunParams         = map[string]string{
		"torch_dtype": "auto",
		"pipeline":    "text-generation",
	}
)

var phiA phi2

type phi2 struct{}

func (*phi2) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ModelFamilyName:           "Phi",
		ImageAccessMode:           string(kaitov1alpha1.ModelImageAccessModePublic),
		DiskStorageRequirement:    "50Gi",
		GPUCountRequirement:       "1",
		TotalGPUMemoryRequirement: "12Gi",
		PerGPUMemoryRequirement:   "0Gi", // We run Phi using native vertical model parallel, no per GPU memory requirement.
		TorchRunParams:            inference.DefaultAccelerateParams,
		ModelRunParams:            phiRunParams,
		ReadinessTimeout:          time.Duration(30) * time.Minute,
		BaseCommand:               baseCommandPresetPhi,
		Tag:                       PresetPhiTagMap["Phi2"],
	}
}
func (*phi2) GetTuningParameters() *model.PresetParam {
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
		Tag:              PresetPhiTagMap["Phi2"],
	}
}
func (*phi2) SupportDistributedInference() bool {
	return false
}
func (*phi2) SupportTuning() bool {
	return true
}
