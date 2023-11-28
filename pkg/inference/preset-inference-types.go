// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package inference

import (
	"fmt"
	"os"
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultNnodes       = "1"
	DefaultNprocPerNode = "1"
	DefaultNodeRank     = "0"
	DefaultMasterAddr   = "localhost"
	DefaultMasterPort   = "29500"
)

// Torch Rendezvous Params
const (
	DefaultMaxRestarts  = "3"
	DefaultRdzvId       = "rdzv_id"
	DefaultRdzvBackend  = "c10d"            // Pytorch Native Distributed data store
	DefaultRdzvEndpoint = "localhost:29500" // e.g. llama-2-13b-chat-0.llama-headless.default.svc.cluster.local:29500
)

const (
	DefaultConfigFile   = "config.yaml"
	DefaultNumProcesses = "1"
	DefaultNumMachines  = "1"
	DefaultMachineRank  = "0"
	DefaultGPUIds       = "all"
)

const (
	DefaultModelId = "MODEL_ID"
	DefaultPipeline = "text-generation"
)

var (
	registryName = os.Getenv("PRESET_REGISTRY_NAME")

	presetFalcon7bImage         = registryName + fmt.Sprintf("/kaito-%s:0.0.1", kaitov1alpha1.PresetFalcon7BModel)
	presetFalcon7bInstructImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", kaitov1alpha1.PresetFalcon7BInstructModel)

	presetFalcon40bImage         = registryName + fmt.Sprintf("/kaito-%s:0.0.1", kaitov1alpha1.PresetFalcon40BModel)
	presetFalcon40bInstructImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", kaitov1alpha1.PresetFalcon40BInstructModel)

	presetMistral7bImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", kaitov1alpha1.PresetMistral7BModel)
	presetMistral7bInstructImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", kaitov1alpha1.PresetMistral7BInstructModel)

	baseCommandPresetLlama = "cd /workspace/llama/llama-2 && torchrun"
	// llamaTextInferenceFile       = "inference-api.py" TODO: To support Text Generation Llama Models
	llamaChatInferenceFile = "inference-api.py"
	llamaRunParams         = map[string]string{
		"max_seq_len":    "512",
		"max_batch_size": "8",
	}

	baseCommandPresetTsf = "accelerate launch --use_deepspeed"
	tsfInferenceFile     = "inference-api.py"
	tsfRunParams         = map[string]string{
		"model_id": DefaultModelId, 
		"pipeline": DefaultPipeline
	}

	defaultTorchRunParams = map[string]string{
		"nnodes":         DefaultNnodes,
		"nproc_per_node": DefaultNprocPerNode,
		"node_rank":      DefaultNodeRank,
		"master_addr":    DefaultMasterAddr,
		"master_port":    DefaultMasterPort,
	}

	defaultTorchRunRdzvParams = map[string]string{
		"max_restarts":  DefaultMaxRestarts,
		"rdzv_id":       DefaultRdzvId,
		"rdzv_backend":  DefaultRdzvBackend,
		"rdzv_endpoint": DefaultRdzvEndpoint,
	}

	defaultAccelerateParams = map[string]string{
		"config_file":   DefaultConfigFile,
		"num_processes": DefaultNumProcesses,
		"num_machines":  DefaultNumMachines,
		"machine_rank":  DefaultMachineRank,
		"gpu_ids":       DefaultGPUIds,
	}

	defaultAccessMode       = "public"
	defaultImagePullSecrets = []corev1.LocalObjectReference{}
)

// PresetInferenceParam defines the preset inference.
type PresetInferenceParam struct {
	ModelName              string
	ModelId				   string
	Image                  string
	ImagePullSecrets       []corev1.LocalObjectReference
	AccessMode             string
	DiskStorageRequirement string
	GPURequirement         string
	GPUMemoryRequirement   string
	TorchRunParams         map[string]string
	TorchRunRdzvParams     map[string]string
	ModelRunParams         map[string]string
	InferenceFile          string
	// DeploymentTimeout defines the maximum duration for pulling the Preset image.
	// This timeout accommodates the size of PresetX, ensuring pull completion
	// even under slower network conditions or unforeseen delays.
	DeploymentTimeout time.Duration
	BaseCommand       string
	// WorldSize defines num of processes required for inference
	WorldSize              int
	DefaultVolumeMountPath string
}

var (

	// Llama2PresetInferences defines the preset inferences for LLaMa2.
	Llama2PresetInferences = map[kaitov1alpha1.ModelName]PresetInferenceParam{

		kaitov1alpha1.PresetLlama2AChat: {
			ModelName:              "LLaMa2",
			Image:                  "",
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "34Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultTorchRunParams,
			TorchRunRdzvParams:     defaultTorchRunRdzvParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(10) * time.Minute,
			BaseCommand:            baseCommandPresetLlama,
			WorldSize:              1,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetLlama2AModel: {
			ModelName:              "LLaMa2",
			Image:                  "",
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "34Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultTorchRunParams,
			TorchRunRdzvParams:     defaultTorchRunRdzvParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(10) * time.Minute,
			BaseCommand:            baseCommandPresetLlama,
			WorldSize:              1,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetLlama2BChat: {
			ModelName:              "LLaMa2",
			Image:                  "",
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "46Gi",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultTorchRunParams,
			TorchRunRdzvParams:     defaultTorchRunRdzvParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(20) * time.Minute,
			BaseCommand:            baseCommandPresetLlama,
			WorldSize:              2,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetLlama2BModel: {
			ModelName:              "LLaMa2",
			Image:                  "",
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "46Gi",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultTorchRunParams,
			TorchRunRdzvParams:     defaultTorchRunRdzvParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(20) * time.Minute,
			BaseCommand:            baseCommandPresetLlama,
			WorldSize:              2,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetLlama2CChat: {
			ModelName:              "LLaMa2",
			Image:                  "",
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "158Gi",
			GPURequirement:         "8",
			GPUMemoryRequirement:   "19Gi",
			TorchRunParams:         defaultTorchRunParams,
			TorchRunRdzvParams:     defaultTorchRunRdzvParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetLlama,
			WorldSize:              8,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetLlama2CModel: {
			ModelName:              "LLaMa2",
			Image:                  "",
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "158Gi",
			GPURequirement:         "8",
			GPUMemoryRequirement:   "19Gi",
			TorchRunParams:         defaultTorchRunParams,
			TorchRunRdzvParams:     defaultTorchRunRdzvParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetLlama,
			WorldSize:              8,
			DefaultVolumeMountPath: "/dev/shm",
		},
	}

	// FalconPresetInferences defines the preset inferences for Falcon.
	FalconPresetInferences = map[kaitov1alpha1.ModelName]PresetInferenceParam{
		kaitov1alpha1.PresetFalcon7BModel: {
			ModelName:              "Falcon",
			ModelId: 				"tiiuae/falcon-7b",
			Image:                  presetFalcon7bImage,
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "50Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "14Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         tsfRunParams,
			InferenceFile:          tsfInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetTsf,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetFalcon7BInstructModel: {
			ModelName:              "Falcon",
			ModelId: 				"tiiuae/falcon-7b-instruct",
			Image:                  presetFalcon7bInstructImage,
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "50Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "14Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         tsfRunParams,
			InferenceFile:          tsfInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetTsf,
			DefaultVolumeMountPath: "/dev/shm",
		},

		kaitov1alpha1.PresetFalcon40BModel: {
			ModelName:              "Falcon",
			ModelId:				"tiiuae/falcon-40b",
			Image:                  presetFalcon40bImage,
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "400",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "90Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         tsfRunParams,
			InferenceFile:          tsfInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetTsf,
			DefaultVolumeMountPath: "/dev/shm",
		},

		kaitov1alpha1.PresetFalcon40BInstructModel: {
			ModelName:              "Falcon",
			ModelId: 				"tiiuae/falcon-40b-instruct",
			Image:                  presetFalcon40bInstructImage,
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "400",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "90Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         tsfRunParams,
			InferenceFile:          tsfInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetTsf,
			DefaultVolumeMountPath: "/dev/shm",
		},
	},

	// MistralPresetInferences defines the preset inferences for Mistral.
	MistralPresetInferences = map[kaitov1alpha1.ModelName]PresetInferenceParam{
		kaitov1alpha1.PresetMistral7BModel: {
			ModelName:              "Mistral",
			ModelId: 				"mistralai/Mistral-7B-v0.1",
			Image:                  presetMistral7bImage,
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "50Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         tsfRunParams,
			InferenceFile:          tsfInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetTsf,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetMistral7BInstructModel: {
			ModelName:              "Mistral",
			ModelId: 				"mistralai/Mistral-7B-Instruct-v0.1",
			Image:                  presetMistral7bInstructImage,
			ImagePullSecrets:       defaultImagePullSecrets,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "50Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         tsfRunParams,
			InferenceFile:          tsfInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetTsf,
			DefaultVolumeMountPath: "/dev/shm",
		},
	}
)
