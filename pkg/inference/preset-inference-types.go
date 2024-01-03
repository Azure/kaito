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

	defaultAccessMode       = "public"
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

	DefaultAccessMode       = "public"
	DefaultImagePullSecrets = []corev1.LocalObjectReference{}
)
