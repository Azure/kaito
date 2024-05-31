// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package model

import (
	"time"
)

type Model interface {
	GetInferenceParameters() *PresetParam
	GetTuningParameters() *PresetParam
	SupportDistributedInference() bool //If true, the model workload will be a StatefulSet, using the torch elastic runtime framework.
	SupportTuning() bool
}

// PresetParam defines the preset inference parameters for a model.
type PresetParam struct {
	ModelFamilyName               string            // The name of the model family.
	ImageAccessMode               string            // Defines where the Image is Public or Private.
	DiskStorageRequirement        string            // Disk storage requirements for the model.
	GPUCountRequirement           string            // Number of GPUs required for the Preset. Used for inference.
	TotalGPUMemoryRequirement     string            // Total GPU memory required for the Preset. Used for inference.
	PerGPUMemoryRequirement       string            // GPU memory required per GPU. Used for inference.
	TuningPerGPUMemoryRequirement map[string]int    // Min GPU memory per tuning method (batch size 1). Used for tuning.
	TorchRunParams                map[string]string // Parameters for configuring the torchrun command.
	TorchRunRdzvParams            map[string]string // Optional rendezvous parameters for distributed training/inference using torchrun (elastic).
	BaseCommand                   string            // The initial command (e.g., 'torchrun', 'accelerate launch') used in the command line.
	ModelRunParams                map[string]string // Parameters for running the model training/inference.
	// ReadinessTimeout defines the maximum duration for creating the workload.
	// This timeout accommodates the size of the image, ensuring pull completion
	// even under slower network conditions or unforeseen delays.
	ReadinessTimeout time.Duration
	WorldSize        int    // Defines the number of processes required for distributed inference.
	Tag              string // The model image tag
}
