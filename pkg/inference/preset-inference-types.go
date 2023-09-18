package inference

import (
	"fmt"
	"os"
	"time"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
)

var (
	registryName = os.Getenv("PRESET_REGISTRY_NAME")

	presetLlama2AChatImage = registryName + fmt.Sprintf("/%s:latest", kdmv1alpha1.PresetLlama2AChat)
	presetLlama2BChatImage = registryName + fmt.Sprintf("/%s:latest", kdmv1alpha1.PresetLlama2BChat)
	presetLlama2CChatImage = registryName + fmt.Sprintf("/%s:latest", kdmv1alpha1.PresetLlama2CChat)

	baseCommandPresetLlama2AChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun web_example_chat_completion.py", kdmv1alpha1.PresetLlama2AChat)
	baseCommandPresetLlama2BChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun --nproc_per_node=2 web_example_chat_completion.py", kdmv1alpha1.PresetLlama2BChat)
	baseCommandPresetLlama2CChat = fmt.Sprintf("cd /workspace/llama/%s && torchrun --nproc_per_node=4 web_example_chat_completion.py", kdmv1alpha1.PresetLlama2CChat)

	torchRunParams = map[string]string{
		"max_seq_len":    "512",
		"max_batch_size": "8",
	}
)

// PresetInferenceParam defines the preset inference.
type PresetInferenceParam struct {
	Image                  string
	DiskStorageRequirement string
	GPURequirement         string
	GPUMemoryRequirement   string
	TorchRunParams         map[string]string
	// DeploymentTimeout defines the maximum duration for pulling the Preset image.
	// This timeout accommodates the size of PresetX, ensuring pull completion
	// even under slower network conditions or unforeseen delays.
	DeploymentTimeout      time.Duration
	BaseCommand            string
	DefaultVolumeMountPath string
}

var (

	// Llama2PresetInferences	defines the preset inferences for LLaMa2.
	Llama2PresetInferences = map[kdmv1alpha1.PresetModelName]PresetInferenceParam{

		kdmv1alpha1.PresetLlama2AChat: {
			Image:                  presetLlama2AChatImage,
			DiskStorageRequirement: "34Gi",
			GPURequirement:         "1",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         torchRunParams,
			DeploymentTimeout:      time.Duration(10) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2AChat,

			DefaultVolumeMountPath: "/dev/shm",
		},
		kdmv1alpha1.PresetLlama2BChat: {
			Image:                  presetLlama2BChatImage,
			DiskStorageRequirement: "46Gi",
			GPURequirement:         "2",
			GPUMemoryRequirement:   "16Gi",
			TorchRunParams:         torchRunParams,
			DeploymentTimeout:      time.Duration(20) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2BChat,

			DefaultVolumeMountPath: "/dev/shm",
		},
		kdmv1alpha1.PresetLlama2CChat: {
			Image:                  presetLlama2CChatImage,
			DiskStorageRequirement: "158Gi",
			GPURequirement:         "4",
			GPUMemoryRequirement:   "19Gi",
			TorchRunParams:         torchRunParams,
			DeploymentTimeout:      time.Duration(30) * time.Minute,
			BaseCommand:            baseCommandPresetLlama2CChat,
			DefaultVolumeMountPath: "/dev/shm",
		},
	}
)
