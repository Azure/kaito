// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package model

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type Model interface {
	GetInferenceParameters() *PresetInferenceParam
	SupportDistributedInference() bool //If true, the model workload will be a StatefulSet, using the torch elastic runtime framework.
}

// PresetInferenceParam defines the preset inference parameters for a model.
type PresetInferenceParam struct {
	ModelFamilyName           string                        // The name of the model family.
	Image                     string                        // Docker image used for running the inference.
	ImagePullSecrets          []corev1.LocalObjectReference // Secrets for pulling the image from a private registry.
	ImageAccessMode           string                        // Defines where the Image is Public or Private.
	DiskStorageRequirement    string                        // Disk storage requirements for the model.
	GPUCountRequirement       string                        // Number of GPUs required for the inference.
	TotalGPUMemoryRequirement string                        // Total GPU memory required for the inference.
	PerGPUMemoryRequirement   string                        // GPU memory required per GPU.
	TorchRunParams            map[string]string             // Parameters for configuring the torchrun command.
	TorchRunRdzvParams        map[string]string             // Optional rendezvous parameters for distributed inference using torchrun (elastic).
	ModelRunParams            map[string]string             // Parameters for running the model inference.
	// DeploymentTimeout defines the maximum duration for pulling the Preset image.
	// This timeout accommodates the size of the image, ensuring pull completion
	// even under slower network conditions or unforeseen delays.
	DeploymentTimeout time.Duration
	// BaseCommand is the initial command (e.g., 'torchrun', 'accelerate launch') used in the command line.
	BaseCommand string
	// WorldSize defines the number of processes required for distributed inference.
	WorldSize              int
}
