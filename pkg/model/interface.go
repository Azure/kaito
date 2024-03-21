// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package model

import (
	"time"
)

type Model interface {
	GetInferenceParameters() *PresetParam
	GetTrainingParameters() *PresetParam
	SupportDistributedInference() bool //If true, the model workload will be a StatefulSet, using the torch elastic runtime framework.
	SupportTraining() bool
}

// PresetParam defines the preset inference parameters for a model.
type PresetParam struct {
	ModelFamilyName           string            // The name of the model family.
	ImageAccessMode           string            // Defines where the Image is Public or Private.
	DiskStorageRequirement    string            // Disk storage requirements for the model.
	GPUCountRequirement       string            // Number of GPUs required for the Preset.
	TotalGPUMemoryRequirement string            // Total GPU memory required for the Preset.
	PerGPUMemoryRequirement   string            // GPU memory required per GPU.
	TorchRunParams            map[string]string // Parameters for configuring the torchrun command.
	TorchRunRdzvParams        map[string]string // Optional rendezvous parameters for distributed training/inference using torchrun (elastic).
	// BaseCommand is the initial command (e.g., 'torchrun', 'accelerate launch') used in the command line.
	BaseCommand    string
	ModelRunParams map[string]string // Parameters for running the model training/inference.
	// WorkloadTimeout defines the maximum duration for creating the workload.
	// This timeout accommodates the size of the image, ensuring pull completion
	// even under slower network conditions or unforeseen delays.
	WorkloadTimeout time.Duration
	// WorldSize defines the number of processes required for distributed inference.
	WorldSize int
	Tag       string // The model image tag
}
