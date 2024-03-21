// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package utils

import (
	"time"

	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils/plugin"
)

type testModel struct{}

func (*testModel) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement: "1",
		WorkloadTimeout:     time.Duration(30) * time.Minute,
	}
}
func (*testModel) GetTrainingParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement: "1",
		WorkloadTimeout:     time.Duration(30) * time.Minute,
	}
}
func (*testModel) SupportDistributedInference() bool {
	return false
}
func (*testModel) SupportTraining() bool {
	return true
}

type testDistributedModel struct{}

func (*testDistributedModel) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement: "1",
		WorkloadTimeout:     time.Duration(30) * time.Minute,
	}
}
func (*testDistributedModel) GetTrainingParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement: "1",
		WorkloadTimeout:     time.Duration(30) * time.Minute,
	}
}
func (*testDistributedModel) SupportDistributedInference() bool {
	return true
}
func (*testDistributedModel) SupportTraining() bool {
	return true
}

func RegisterTestModel() {
	var test testModel
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "test-model",
		Instance: &test,
	})

	var testDistributed testDistributedModel
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "test-distributed-model",
		Instance: &testDistributed,
	})

}
