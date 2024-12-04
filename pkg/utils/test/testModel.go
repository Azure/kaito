// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package test

import (
	"time"

	"github.com/kaito-project/kaito/pkg/model"
	"github.com/kaito-project/kaito/pkg/utils/plugin"
)

type baseTestModel struct{}

func (*baseTestModel) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement: "1",
		RuntimeParam: model.RuntimeParam{
			VLLM: model.VLLMParam{
				BaseCommand: "python3 /workspace/vllm/inference_api.py",
				ModelName:   "mymodel",
			},
			Transformers: model.HuggingfaceTransformersParam{
				BaseCommand:       "accelerate launch",
				InferenceMainFile: "/workspace/tfs/inference_api.py",
			},
		},
		ReadinessTimeout: time.Duration(30) * time.Minute,
	}
}
func (*baseTestModel) GetTuningParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement: "1",
		ReadinessTimeout:    time.Duration(30) * time.Minute,
	}
}
func (*baseTestModel) SupportDistributedInference() bool {
	return true
}
func (*baseTestModel) SupportTuning() bool {
	return true
}

type testModel struct {
	baseTestModel
}

func (*testModel) SupportDistributedInference() bool {
	return false
}

type testDistributedModel struct {
	baseTestModel
}

type testNoTensorParallelModel struct {
	baseTestModel
}

func (*testNoTensorParallelModel) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement: "1",
		RuntimeParam: model.RuntimeParam{
			VLLM: model.VLLMParam{
				BaseCommand:               "python3 /workspace/vllm/inference_api.py",
				TensorParallelUnsupported: true,
			},
			Transformers: model.HuggingfaceTransformersParam{
				BaseCommand:       "accelerate launch",
				InferenceMainFile: "/workspace/tfs/inference_api.py",
			},
		},
		ReadinessTimeout: time.Duration(30) * time.Minute,
	}
}
func (*testNoTensorParallelModel) SupportDistributedInference() bool {
	return false
}

func RegisterTestModel() {
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "test-model",
		Instance: &testModel{},
	})

	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "test-distributed-model",
		Instance: &testDistributedModel{},
	})

	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "test-no-tensor-parallel-model",
		Instance: &testNoTensorParallelModel{},
	})
}
