// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package falcon

import (
	"fmt"
	"os"
	"time"

	"github.com/azure/kaito/pkg/inference"
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
	falconInferenceFile     = "inference-api.py"
	falconRunParams         = map[string]string{}
)

var falconA falcon7b

type falcon7b struct{}

func (*falcon7b) GetInferenceParameters() *inference.PresetInferenceParam {
	return &inference.PresetInferenceParam{
		ModelName:              "Falcon",
		Image:                  presetFalcon7bImage,
		ImagePullSecrets:       inference.DefaultImagePullSecrets,
		AccessMode:             inference.DefaultAccessMode,
		DiskStorageRequirement: "50Gi",
		GPURequirement:         "1",
		GPUMemoryRequirement:   "14Gi",
		TorchRunParams:         inference.DefaultAccelerateParams,
		ModelRunParams:         falconRunParams,
		InferenceFile:          falconInferenceFile,
		DeploymentTimeout:      time.Duration(30) * time.Minute,
		BaseCommand:            baseCommandPresetFalcon,
		DefaultVolumeMountPath: "/dev/shm",
	}

}
func (*falcon7b) NeedStatefulSet() bool {
	return false
}
func (*falcon7b) NeedHeadlessService() bool {
	return false
}

var falconB falcon7bInst

type falcon7bInst struct{}

func (*falcon7bInst) GetInferenceParameters() *inference.PresetInferenceParam {
	return &inference.PresetInferenceParam{
		ModelName:              "Falcon",
		Image:                  presetFalcon7bInstructImage,
		ImagePullSecrets:       inference.DefaultImagePullSecrets,
		AccessMode:             inference.DefaultAccessMode,
		DiskStorageRequirement: "50Gi",
		GPURequirement:         "1",
		GPUMemoryRequirement:   "14Gi",
		TorchRunParams:         inference.DefaultAccelerateParams,
		ModelRunParams:         falconRunParams,
		InferenceFile:          falconInferenceFile,
		DeploymentTimeout:      time.Duration(30) * time.Minute,
		BaseCommand:            baseCommandPresetFalcon,
		DefaultVolumeMountPath: "/dev/shm",
	}

}
func (*falcon7bInst) NeedStatefulSet() bool {
	return false
}
func (*falcon7bInst) NeedHeadlessService() bool {
	return false
}

var falconC falcon40b

type falcon40b struct{}

func (*falcon40b) GetInferenceParameters() *inference.PresetInferenceParam {
	return &inference.PresetInferenceParam{
		ModelName:              "Falcon",
		Image:                  presetFalcon40bImage,
		ImagePullSecrets:       inference.DefaultImagePullSecrets,
		AccessMode:             inference.DefaultAccessMode,
		DiskStorageRequirement: "400",
		GPURequirement:         "2",
		GPUMemoryRequirement:   "90Gi",
		TorchRunParams:         inference.DefaultAccelerateParams,
		ModelRunParams:         falconRunParams,
		InferenceFile:          falconInferenceFile,
		DeploymentTimeout:      time.Duration(30) * time.Minute,
		BaseCommand:            baseCommandPresetFalcon,
		DefaultVolumeMountPath: "/dev/shm",
	}

}
func (*falcon40b) NeedStatefulSet() bool {
	return false
}
func (*falcon40b) NeedHeadlessService() bool {
	return false
}

var falconD falcon40bInst

type falcon40bInst struct{}

func (*falcon40bInst) GetInferenceParameters() *inference.PresetInferenceParam {
	return &inference.PresetInferenceParam{
		ModelName:              "Falcon",
		Image:                  presetFalcon40bInstructImage,
		ImagePullSecrets:       inference.DefaultImagePullSecrets,
		AccessMode:             inference.DefaultAccessMode,
		DiskStorageRequirement: "400",
		GPURequirement:         "2",
		GPUMemoryRequirement:   "90Gi",
		TorchRunParams:         inference.DefaultAccelerateParams,
		ModelRunParams:         falconRunParams,
		InferenceFile:          falconInferenceFile,
		DeploymentTimeout:      time.Duration(30) * time.Minute,
		BaseCommand:            baseCommandPresetFalcon,
		DefaultVolumeMountPath: "/dev/shm",
	}

}
func (*falcon40bInst) NeedStatefulSet() bool {
	return false
}
func (*falcon40bInst) NeedHeadlessService() bool {
	return false
}
