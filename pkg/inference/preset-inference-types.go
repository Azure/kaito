package inference

import (
	"fmt"
	"os"
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
)

const (
	DefaultNnodes       = "1"
	DefaultNprocPerNode = "1"
	DefaultNodeRank     = "0"
	DefaultMasterAddr   = "localhost"
	DefaultMasterPort   = "29500"
)

const (
	DefaultConfigFile   = "config.yaml"
	DefaultNumProcesses = "1"
	DefaultNumMachines  = "1"
	DefaultMachineRank  = "0"
	DefaultGPUIds       = "all"
)

var (
	registryName = os.Getenv("PRESET_REGISTRY_NAME")

	presetLlama2AChatImage = registryName + fmt.Sprintf("/%s:latest", kaitov1alpha1.PresetLlama2AChat)
	presetLlama2BChatImage = registryName + fmt.Sprintf("/%s:latest", kaitov1alpha1.PresetLlama2BChat)
	presetLlama2CChatImage = registryName + fmt.Sprintf("/%s:latest", kaitov1alpha1.PresetLlama2CChat)

	presetFalcon7bImage         = registryName + fmt.Sprintf("/%s:latest", kaitov1alpha1.PresetFalcon7BModel)
	presetFalcon7bInstructImage = registryName + fmt.Sprintf("/%s:latest", kaitov1alpha1.PresetFalcon7BInstructModel)

	presetFalcon40bImage         = registryName + fmt.Sprintf("/%s:latest", kaitov1alpha1.PresetFalcon40BModel)
	presetFalcon40bInstructImage = registryName + fmt.Sprintf("/%s:latest", kaitov1alpha1.PresetFalcon40BInstructModel)

	baseCommandPresetLlama2AChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun", kaitov1alpha1.PresetLlama2AChat)
	baseCommandPresetLlama2BChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun", kaitov1alpha1.PresetLlama2BChat)
	baseCommandPresetLlama2CChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun", kaitov1alpha1.PresetLlama2CChat)
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

	defaultAccelerateParams = map[string]string{
		"config_file":   DefaultConfigFile,
		"num_processes": DefaultNumProcesses,
		"num_machines":  DefaultNumMachines,
		"machine_rank":  DefaultMachineRank,
		"gpu_ids":       DefaultGPUIds,
	}

	defaultAccessMode = "public"
)

// PresetInferenceParam defines the preset inference.
type PresetInferenceParam struct {
	ModelName              string
	Image                  string
	AccessMode             string
	DiskStorageRequirement string
	GPURequirement         string
	GPUMemoryRequirement   string
	TorchRunParams         map[string]string
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
			Image:                  presetLlama2AChatImage,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "34Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultTorchRunParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(10) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2AChat,
			WorldSize:              1,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetLlama2BChat: {
			ModelName:              "LLaMa2",
			Image:                  presetLlama2BChatImage,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "46Gi",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultTorchRunParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(20) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2BChat,
			WorldSize:              2,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetLlama2CChat: {
			ModelName:              "LLaMa2",
			Image:                  presetLlama2CChatImage,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "158Gi",
			GPURequirement:         "8",
			GPUMemoryRequirement:   "19Gi",
			TorchRunParams:         defaultTorchRunParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaChatInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2CChat,
			WorldSize:              8,
			DefaultVolumeMountPath: "/dev/shm",
		},
	}

	// FalconPresetInferences defines the preset inferences for Falcon.
	FalconPresetInferences = map[kaitov1alpha1.ModelName]PresetInferenceParam{
		kaitov1alpha1.PresetFalcon7BModel: {
			ModelName:              "Falcon",
			Image:                  presetFalcon7bImage,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "50Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "14Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         falconRunParams,
			InferenceFile:          falconInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetFalcon,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kaitov1alpha1.PresetFalcon7BInstructModel: {
			ModelName:              "Falcon",
			Image:                  presetFalcon7bInstructImage,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "50Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "14Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         falconRunParams,
			InferenceFile:          falconInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetFalcon,
			DefaultVolumeMountPath: "/dev/shm",
		},

		kaitov1alpha1.PresetFalcon40BModel: {
			ModelName:              "Falcon",
			Image:                  presetFalcon40bImage,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "400",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "90Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         falconRunParams,
			InferenceFile:          falconInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetFalcon,
			DefaultVolumeMountPath: "/dev/shm",
		},

		kaitov1alpha1.PresetFalcon40BInstructModel: {
			ModelName:              "Falcon",
			Image:                  presetFalcon40bInstructImage,
			AccessMode:             defaultAccessMode,
			DiskStorageRequirement: "400",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "90Gi",
			TorchRunParams:         defaultAccelerateParams,
			ModelRunParams:         falconRunParams,
			InferenceFile:          falconInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetFalcon,
			DefaultVolumeMountPath: "/dev/shm",
		},
	}
)
