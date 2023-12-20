// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package inference

import (
	"fmt"
	"os"
	"time"

	"github.com/azure/kaito/pkg/model"
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
	PresetLlama2AModel = "llama-2-7b"
	PresetLlama2BModel = "llama-2-13b"
	PresetLlama2CModel = "llama-2-70b"
	PresetLlama2AChat  = PresetLlama2AModel + "-chat"
	PresetLlama2BChat  = PresetLlama2BModel + "-chat"
	PresetLlama2CChat  = PresetLlama2CModel + "-chat"

	PresetFalcon7BModel          = "falcon-7b"
	PresetFalcon40BModel         = "falcon-40b"
	PresetFalcon7BInstructModel  = PresetFalcon7BModel + "-instruct"
	PresetFalcon40BInstructModel = PresetFalcon40BModel + "-instruct"
)

var (
	registryName = os.Getenv("PRESET_REGISTRY_NAME")

	presetFalcon7bImage         = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetFalcon7BModel)
	presetFalcon7bInstructImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetFalcon7BInstructModel)

	presetFalcon40bImage         = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetFalcon40BModel)
	presetFalcon40bInstructImage = registryName + fmt.Sprintf("/kaito-%s:0.0.1", PresetFalcon40BInstructModel)

	baseCommandPresetLlama = "cd /workspace/llama/llama-2 && torchrun"
	// llamaTextInferenceFile       = "inference-api.py" TODO: To support Text Generation Llama Models
	llamaChatInferenceFile = "inference-api.py"
	llamaRunParams         = map[string]string{
		"max_seq_len":    "512",
		"max_batch_size": "8",
	}

	baseCommandPresetFalcon = "accelerate launch --use_deepspeed"
	falconInferenceFile     = "inference-api.py"
	falconRunParams         = map[string]string{}

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

	defaultImageAccessMode  = "public"
	defaultImagePullSecrets = []corev1.LocalObjectReference{}
)

// TODO: remove the above local variables starting with lower cases.
var (
	DefaultTorchRunParams = map[string]string{
		"nnodes":         DefaultNnodes,
		"nproc_per_node": DefaultNprocPerNode,
		"node_rank":      DefaultNodeRank,
		"master_addr":    DefaultMasterAddr,
		"master_port":    DefaultMasterPort,
	}

	DefaultTorchRunRdzvParams = map[string]string{
		"max_restarts":  DefaultMaxRestarts,
		"rdzv_id":       DefaultRdzvId,
		"rdzv_backend":  DefaultRdzvBackend,
		"rdzv_endpoint": DefaultRdzvEndpoint,
	}

	DefaultAccelerateParams = map[string]string{
		"config_file":   DefaultConfigFile,
		"num_processes": DefaultNumProcesses,
		"num_machines":  DefaultNumMachines,
		"machine_rank":  DefaultMachineRank,
		"gpu_ids":       DefaultGPUIds,
	}

	DefaultImageAccessMode  = "public"
	DefaultImagePullSecrets = []corev1.LocalObjectReference{}
)

var (

	// Llama2PresetInferences defines the preset inferences for LLaMa2.
	Llama2PresetInferences = map[string]model.PresetInferenceParam{

		PresetLlama2AChat: {
			ModelFamilyName:           "LLaMa2",
			Image:                     "",
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultImageAccessMode,
			DiskStorageRequirement:    "34Gi",
			GPUCountRequirement:       "1",
			TotalGPUMemoryRequirement: "16Gi",
			TorchRunParams:            defaultTorchRunParams,
			TorchRunRdzvParams:        defaultTorchRunRdzvParams,
			ModelRunParams:            llamaRunParams,
			InferenceFile:             llamaChatInferenceFile,
			DeploymentTimeout:         time.Duration(10) * time.Minute,
			BaseCommand:               baseCommandPresetLlama,
			WorldSize:                 1,
			DefaultVolumeMountPath:    "/dev/shm",
		},
		PresetLlama2AModel: {
			ModelFamilyName:           "LLaMa2",
			Image:                     "",
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultAccessMode,
			DiskStorageRequirement:    "34Gi",
			GPUCountRequirement:       "1",
			TotalGPUMemoryRequirement: "16Gi",
			TorchRunParams:            defaultTorchRunParams,
			TorchRunRdzvParams:        defaultTorchRunRdzvParams,
			ModelRunParams:            llamaRunParams,
			InferenceFile:             llamaChatInferenceFile,
			DeploymentTimeout:         time.Duration(10) * time.Minute,
			BaseCommand:               baseCommandPresetLlama,
			WorldSize:                 1,
			DefaultVolumeMountPath:    "/dev/shm",
		},
		PresetLlama2BChat: {
			ModelFamilyName:           "LLaMa2",
			Image:                     "",
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultAccessMode,
			DiskStorageRequirement:    "46Gi",
			GPUCountRequirement:       "2",
			TotalGPUMemoryRequirement: "16Gi",
			TorchRunParams:            defaultTorchRunParams,
			TorchRunRdzvParams:        defaultTorchRunRdzvParams,
			ModelRunParams:            llamaRunParams,
			InferenceFile:             llamaChatInferenceFile,
			DeploymentTimeout:         time.Duration(20) * time.Minute,
			BaseCommand:               baseCommandPresetLlama,
			WorldSize:                 2,
			DefaultVolumeMountPath:    "/dev/shm",
		},
		PresetLlama2BModel: {
			ModelFamilyName:           "LLaMa2",
			Image:                     "",
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultAccessMode,
			DiskStorageRequirement:    "46Gi",
			GPUCountRequirement:       "2",
			TotalGPUMemoryRequirement: "16Gi",
			TorchRunParams:            defaultTorchRunParams,
			TorchRunRdzvParams:        defaultTorchRunRdzvParams,
			ModelRunParams:            llamaRunParams,
			InferenceFile:             llamaChatInferenceFile,
			DeploymentTimeout:         time.Duration(20) * time.Minute,
			BaseCommand:               baseCommandPresetLlama,
			WorldSize:                 2,
			DefaultVolumeMountPath:    "/dev/shm",
		},
		PresetLlama2CChat: {
			ModelFamilyName:           "LLaMa2",
			Image:                     "",
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultAccessMode,
			DiskStorageRequirement:    "158Gi",
			GPUCountRequirement:       "8",
			TotalGPUMemoryRequirement: "19Gi",
			TorchRunParams:            defaultTorchRunParams,
			TorchRunRdzvParams:        defaultTorchRunRdzvParams,
			ModelRunParams:            llamaRunParams,
			InferenceFile:             llamaChatInferenceFile,
			DeploymentTimeout:         time.Duration(30) * time.Minute,
			BaseCommand:               baseCommandPresetLlama,
			WorldSize:                 8,
			DefaultVolumeMountPath:    "/dev/shm",
		},
		PresetLlama2CModel: {
			ModelFamilyName:           "LLaMa2",
			Image:                     "",
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultAccessMode,
			DiskStorageRequirement:    "158Gi",
			GPUCountRequirement:       "8",
			TotalGPUMemoryRequirement: "19Gi",
			TorchRunParams:            defaultTorchRunParams,
			TorchRunRdzvParams:        defaultTorchRunRdzvParams,
			ModelRunParams:            llamaRunParams,
			InferenceFile:             llamaChatInferenceFile,
			DeploymentTimeout:         time.Duration(30) * time.Minute,
			BaseCommand:               baseCommandPresetLlama,
			WorldSize:                 8,
			DefaultVolumeMountPath:    "/dev/shm",
		},
	}

	// FalconPresetInferences defines the preset inferences for Falcon.
	FalconPresetInferences = map[string]model.PresetInferenceParam{
		PresetFalcon7BModel: {
			ModelFamilyName:           "Falcon",
			Image:                     presetFalcon7bImage,
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultImageAccessMode,
			DiskStorageRequirement:    "50Gi",
			GPUCountRequirement:       "1",
			TotalGPUMemoryRequirement: "14Gi",
			TorchRunParams:            defaultAccelerateParams,
			ModelRunParams:            falconRunParams,
			InferenceFile:             falconInferenceFile,
			DeploymentTimeout:         time.Duration(30) * time.Minute,
			BaseCommand:               baseCommandPresetFalcon,
			DefaultVolumeMountPath:    "/dev/shm",
		},
		PresetFalcon7BInstructModel: {
			ModelFamilyName:           "Falcon",
			Image:                     presetFalcon7bInstructImage,
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultImageAccessMode,
			DiskStorageRequirement:    "50Gi",
			GPUCountRequirement:       "1",
			TotalGPUMemoryRequirement: "14Gi",
			TorchRunParams:            defaultAccelerateParams,
			ModelRunParams:            falconRunParams,
			InferenceFile:             falconInferenceFile,
			DeploymentTimeout:         time.Duration(30) * time.Minute,
			BaseCommand:               baseCommandPresetFalcon,
			DefaultVolumeMountPath:    "/dev/shm",
		},

		PresetFalcon40BModel: {
			ModelFamilyName:           "Falcon",
			Image:                     presetFalcon40bImage,
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultImageAccessMode,
			DiskStorageRequirement:    "400",
			GPUCountRequirement:       "2",
			TotalGPUMemoryRequirement: "90Gi",
			TorchRunParams:            defaultAccelerateParams,
			ModelRunParams:            falconRunParams,
			InferenceFile:             falconInferenceFile,
			DeploymentTimeout:         time.Duration(30) * time.Minute,
			BaseCommand:               baseCommandPresetFalcon,
			DefaultVolumeMountPath:    "/dev/shm",
		},

		PresetFalcon40BInstructModel: {
			ModelFamilyName:           "Falcon",
			Image:                     presetFalcon40bInstructImage,
			ImagePullSecrets:          defaultImagePullSecrets,
			ImageAccessMode:           defaultImageAccessMode,
			DiskStorageRequirement:    "400",
			GPUCountRequirement:       "2",
			TotalGPUMemoryRequirement: "90Gi",
			TorchRunParams:            defaultAccelerateParams,
			ModelRunParams:            falconRunParams,
			InferenceFile:             falconInferenceFile,
			DeploymentTimeout:         time.Duration(30) * time.Minute,
			BaseCommand:               baseCommandPresetFalcon,
			DefaultVolumeMountPath:    "/dev/shm",
		},
	}
)
