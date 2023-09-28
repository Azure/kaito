package inference

import (
	"fmt"
	"os"
	"time"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
)

const (
	DefaultNnodes       = "1"
	DefaultNprocPerNode = "1"
	DefaultNodeRank     = "0"
	DefaultMasterAddr   = "localhost"
	DefaultMasterPort   = "29500"
)

var (
	registryName = os.Getenv("PRESET_REGISTRY_NAME")

	presetLlama2AChatImage = registryName + fmt.Sprintf("/%s:latest", kdmv1alpha1.PresetLlama2AChat)
	presetLlama2BChatImage = registryName + fmt.Sprintf("/%s:latest", kdmv1alpha1.PresetLlama2BChat)
	presetLlama2CChatImage = registryName + fmt.Sprintf("/%s:latest", kdmv1alpha1.PresetLlama2CChat)

	baseCommandPresetLlama2AChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun", kdmv1alpha1.PresetLlama2AChat)
	baseCommandPresetLlama2BChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun", kdmv1alpha1.PresetLlama2BChat)
	baseCommandPresetLlama2CChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun", kdmv1alpha1.PresetLlama2CChat)
	llamaInferenceFile           = "web_example_chat_completion.py"
	llamaRunParams               = map[string]string{
		"max_seq_len":    "512",
		"max_batch_size": "8",
	}

	defaultTorchRunParams = map[string]string{
		"nnodes":         DefaultNnodes,
		"nproc_per_node": DefaultNprocPerNode,
		"node_rank":      DefaultNodeRank,
		"master_addr":    DefaultMasterAddr,
		"master_port":    DefaultMasterPort,
	}
)

// PresetInferenceParam defines the preset inference.
type PresetInferenceParam struct {
	ModelName              string
	Image                  string
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
	Llama2PresetInferences = map[kdmv1alpha1.ModelName]PresetInferenceParam{

		kdmv1alpha1.PresetLlama2AChat: {
			ModelName:              "LLaMa2",
			Image:                  presetLlama2AChatImage,
			DiskStorageRequirement: "34Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultTorchRunParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaInferenceFile,
			DeploymentTimeout:      time.Duration(10) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2AChat,
			WorldSize:              1,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kdmv1alpha1.PresetLlama2BChat: {
			ModelName:              "LLaMa2",
			Image:                  presetLlama2BChatImage,
			DiskStorageRequirement: "46Gi",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         defaultTorchRunParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaInferenceFile,
			DeploymentTimeout:      time.Duration(20) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2BChat,
			WorldSize:              2,
			DefaultVolumeMountPath: "/dev/shm",
		},
		kdmv1alpha1.PresetLlama2CChat: {
			ModelName:              "LLaMa2",
			Image:                  presetLlama2CChatImage,
			DiskStorageRequirement: "158Gi",
			GPURequirement:         "8",
			GPUMemoryRequirement:   "19Gi",
			TorchRunParams:         defaultTorchRunParams,
			ModelRunParams:         llamaRunParams,
			InferenceFile:          llamaInferenceFile,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2CChat,
			WorldSize:              8,
			DefaultVolumeMountPath: "/dev/shm",
		},
	}
)
