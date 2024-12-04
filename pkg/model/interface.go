// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package model

import (
	"time"

	"github.com/kaito-project/kaito/pkg/utils"
)

type Model interface {
	GetInferenceParameters() *PresetParam
	GetTuningParameters() *PresetParam
	SupportDistributedInference() bool //If true, the model workload will be a StatefulSet, using the torch elastic runtime framework.
	SupportTuning() bool
}

// RuntimeName is LLM runtime name.
type RuntimeName string

const (
	RuntimeNameHuggingfaceTransformers RuntimeName = "transformers"
	RuntimeNameVLLM                    RuntimeName = "vllm"
)

// PresetParam defines the preset inference parameters for a model.
type PresetParam struct {
	Tag             string // The model image tag
	ModelFamilyName string // The name of the model family.
	ImageAccessMode string // Defines where the Image is Public or Private.

	DiskStorageRequirement        string         // Disk storage requirements for the model.
	GPUCountRequirement           string         // Number of GPUs required for the Preset. Used for inference.
	TotalGPUMemoryRequirement     string         // Total GPU memory required for the Preset. Used for inference.
	PerGPUMemoryRequirement       string         // GPU memory required per GPU. Used for inference.
	TuningPerGPUMemoryRequirement map[string]int // Min GPU memory per tuning method (batch size 1). Used for tuning.
	WorldSize                     int            // Defines the number of processes required for distributed inference.

	RuntimeParam

	// ReadinessTimeout defines the maximum duration for creating the workload.
	// This timeout accommodates the size of the image, ensuring pull completion
	// even under slower network conditions or unforeseen delays.
	ReadinessTimeout time.Duration
}

// RuntimeParam defines the llm runtime parameters.
type RuntimeParam struct {
	Transformers HuggingfaceTransformersParam
	VLLM         VLLMParam
}

type HuggingfaceTransformersParam struct {
	BaseCommand        string            // The initial command (e.g., 'torchrun', 'accelerate launch') used in the command line.
	TorchRunParams     map[string]string // Parameters for configuring the torchrun command.
	TorchRunRdzvParams map[string]string // Optional rendezvous parameters for distributed training/inference using torchrun (elastic).
	InferenceMainFile  string            // The main file for inference.
	ModelRunParams     map[string]string // Parameters for running the model training/inference.
}

type VLLMParam struct {
	BaseCommand string
	// The model name used in the openai serving API.
	// see https://platform.openai.com/docs/api-reference/chat/create#chat-create-model.
	ModelName string
	// Parameters for distributed inference.
	DistributionParams map[string]string
	// Parameters for running the model training/inference.
	ModelRunParams map[string]string
	// Wether the model supports multi-GPU (tensor parallel inference).
	// see https://docs.vllm.ai/en/latest/serving/distributed_serving.html#details-for-distributed-inference-and-serving.
	TensorParallelUnsupported bool
}

func (p *PresetParam) DeepCopy() *PresetParam {
	if p == nil {
		return nil
	}
	out := new(PresetParam)
	*out = *p
	out.RuntimeParam = p.RuntimeParam.DeepCopy()
	out.TuningPerGPUMemoryRequirement = make(map[string]int, len(p.TuningPerGPUMemoryRequirement))
	for k, v := range p.TuningPerGPUMemoryRequirement {
		out.TuningPerGPUMemoryRequirement[k] = v
	}
	return out
}

func (rp *RuntimeParam) DeepCopy() RuntimeParam {
	if rp == nil {
		return RuntimeParam{}
	}
	out := RuntimeParam{}
	out.Transformers = rp.Transformers.DeepCopy()
	out.VLLM = rp.VLLM.DeepCopy()
	return out
}

func (h *HuggingfaceTransformersParam) DeepCopy() HuggingfaceTransformersParam {
	if h == nil {
		return HuggingfaceTransformersParam{}
	}
	out := HuggingfaceTransformersParam{}
	out = *h
	out.TorchRunParams = make(map[string]string, len(h.TorchRunParams))
	for k, v := range h.TorchRunParams {
		out.TorchRunParams[k] = v
	}
	out.TorchRunRdzvParams = make(map[string]string, len(h.TorchRunRdzvParams))
	for k, v := range h.TorchRunRdzvParams {
		out.TorchRunRdzvParams[k] = v
	}
	out.ModelRunParams = make(map[string]string, len(h.ModelRunParams))
	for k, v := range h.ModelRunParams {
		out.ModelRunParams[k] = v
	}
	return out
}

func (v *VLLMParam) DeepCopy() VLLMParam {
	if v == nil {
		return VLLMParam{}
	}
	out := VLLMParam{}
	out = *v
	out.DistributionParams = make(map[string]string, len(v.DistributionParams))
	for k, v := range v.DistributionParams {
		out.DistributionParams[k] = v
	}
	out.ModelRunParams = make(map[string]string, len(v.ModelRunParams))
	for k, v := range v.ModelRunParams {
		out.ModelRunParams[k] = v
	}
	return out
}

// builds the container command:
// eg. torchrun <TORCH_PARAMS> <OPTIONAL_RDZV_PARAMS> baseCommand <MODEL_PARAMS>
func (p *PresetParam) GetInferenceCommand(runtime RuntimeName, skuNumGPUs string) []string {
	switch runtime {
	case RuntimeNameHuggingfaceTransformers:
		torchCommand := utils.BuildCmdStr(p.Transformers.BaseCommand, p.Transformers.TorchRunParams, p.Transformers.TorchRunRdzvParams)
		modelCommand := utils.BuildCmdStr(p.Transformers.InferenceMainFile, p.Transformers.ModelRunParams)
		return utils.ShellCmd(torchCommand + " " + modelCommand)
	case RuntimeNameVLLM:
		if p.VLLM.ModelName != "" {
			p.VLLM.ModelRunParams["served-model-name"] = p.VLLM.ModelName
		}
		if !p.VLLM.TensorParallelUnsupported {
			p.VLLM.ModelRunParams["tensor-parallel-size"] = skuNumGPUs
		}
		modelCommand := utils.BuildCmdStr(p.VLLM.BaseCommand, p.VLLM.ModelRunParams)
		return utils.ShellCmd(modelCommand)
	default:
		return nil
	}
}
