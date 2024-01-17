// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package utils

import (
	"time"

	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils/plugin"
)

type testModel struct{}

func (*testModel) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		GPUCountRequirement: "1",
		DeploymentTimeout:   time.Duration(30) * time.Minute,
		InferenceFile:       "inference-api.py",
	}
}
func (*testModel) SupportDistributedInference() bool {
	return false
}

type testDistributedModel struct{}

func (*testDistributedModel) GetInferenceParameters() *model.PresetInferenceParam {
	return &model.PresetInferenceParam{
		GPUCountRequirement: "1",
		DeploymentTimeout:   time.Duration(30) * time.Minute,
		InferenceFile:       "inference-api.py",
	}
}
func (*testDistributedModel) SupportDistributedInference() bool {
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
