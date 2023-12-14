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

// PresetInferenceParam defines the preset inference.
type PresetInferenceParam struct {
	ModelName                 string
	Image                     string
	ImagePullSecrets          []corev1.LocalObjectReference
	AccessMode                string
	DiskStorageRequirement    string
	GPUCountRequirement       string
	TotalGPUMemoryRequirement string
	PerGPUMemoryRequirement   string
	TorchRunParams            map[string]string
	TorchRunRdzvParams        map[string]string
	ModelRunParams            map[string]string
	InferenceFile             string
	// DeploymentTimeout defines the maximum duration for pulling the Preset image.
	// This timeout accommodates the size of PresetX, ensuring pull completion
	// even under slower network conditions or unforeseen delays.
	DeploymentTimeout time.Duration
	BaseCommand       string
	// WorldSize defines num of processes required for inference
	WorldSize              int
	DefaultVolumeMountPath string
}
