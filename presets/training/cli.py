from dataclasses import asdict, dataclass, field
from typing import Any, Dict, List, Literal, Optional, Union

import torch
from peft import LoraConfig
from transformers import BitsAndBytesConfig

# DEFAULT_LORA_CONFIG = LoraConfig(
#     r=16,
#     lora_alpha=32,
#     target_modules=["query_key_value"],
#     lora_dropout=0.05,
#     bias="none",
#     task_type="CAUSAL_LM"
# )

# TODO: Need to fix this up to allow multiple types for these fields as intended
@dataclass
class ExtLoraConfig(LoraConfig):
    init_lora_weights: bool = field(default=True, metadata={"help": "Enable initialization of LoRA weights"})
    layers_to_transform: Optional[List[int]] = field(default=None, metadata={"help": "Layer indices to apply LoRA"})
    layers_pattern: Optional[List[str]] = field(default=None, metadata={"help": "Pattern to match layers for LoRA"})
    loftq_config: Dict[str, any] = field(default_factory=dict, metadata={"help": "LoftQ configuration for quantization"}) 
#     r: int = field(default=8, metadata={"help": "Lora attention dimension"})
#     target_modules: Optional[Union[List[str], str]] = field(
#         default=None,
#         metadata={"help": "List of module names or regex expressions to replace with LoRA"}
#     )
#     task_type: Optional[str] = field(default=None, metadata={"help": "Megatron core module"})
#     lora_alpha: int = field(default=8, metadata={"help": "Lora alpha scaling factor"})
#     lora_dropout: float = field(default=0.0, metadata={"help": "Dropout rate for LoRA layers"})
#     fan_in_fan_out: bool = field(default=False, metadata={"help": "Set true if layer weights are (fan_in, fan_out)"})
#     bias: Literal["none", "all", "lora_only"] = field(default="none", metadata={"help": "Bias type for LoRA layers"})
#     use_rslora: bool = field(default=False, metadata={"help": "Use Rank-Stabilized LoRA"})
#     modules_to_save: Optional[List[str]] = field(default_factory=list, metadata={"help": "Modules to save besides LoRA layers"})
#     # init_lora_weights: Optional[Union[bool, Literal["gaussian", "loftq"]]] = field(default=True, metadata={"help": "Initialize LoRA weights method"})
#     # layers_to_transform: Optional[Union[List[int], int]] = field(default=None, metadata={"help": "Layer indices to apply LoRA"})
#     # layers_pattern: Optional[Union[List[str], str]] = field(default=None, metadata={"help": "Pattern to match layers for LoRA"})
#     rank_pattern: Optional[Dict[str, int]] = field(default_factory=dict, metadata={"help": "Custom ranks for specific layers"})
#     alpha_pattern: Optional[Dict[str, int]] = field(default_factory=dict, metadata={"help": "Custom alphas for specific layers"})
#     megatron_config: Optional[Dict[str, any]] = field(default=None, metadata={"help": "Megatron TransformerConfig for parallel layers"})
#     megatron_core: Optional[str] = field(default="megatron.core", metadata={"help": "Megatron core module"})
#     # loftq_config: Union[LoftQConfig, Dict[str, any]] = field(default_factory=dict, metadata={"help": "LoftQ configuration for quantization"}) 

@dataclass
class ModelConfig:
    """
    Transformers Model Configuration Parameters
    """
    pretrained_model_name_or_path: Optional[str] = field(default="/workspace/tfs/weights", metadata={"help": "Path to the pretrained model or model identifier from huggingface.co/models"})
    state_dict: Optional[Dict[str, Any]] = field(default=None, metadata={"help": "State dictionary for the model"})
    cache_dir: Optional[str] = field(default=None, metadata={"help": "Cache directory for the model"})
    from_tf: bool = field(default=False, metadata={"help": "Load model from a TensorFlow checkpoint"})
    force_download: bool = field(default=False, metadata={"help": "Force the download of the model"})
    resume_download: bool = field(default=False, metadata={"help": "Resume an interrupted download"})
    proxies: Optional[str] = field(default=None, metadata={"help": "Proxy configuration for downloading the model"})
    output_loading_info: bool = field(default=False, metadata={"help": "Output additional loading information"})
    allow_remote_files: bool = field(default=False, metadata={"help": "Allow using remote files, default is local only"})
    m_revision: str = field(default="main", metadata={"help": "Specific model version to use"})
    trust_remote_code: bool = field(default=False, metadata={"help": "Enable trusting remote code when loading the model"})
    m_load_in_4bit: bool = field(default=False, metadata={"help": "Load model in 4-bit mode"})
    m_load_in_8bit: bool = field(default=False, metadata={"help": "Load model in 8-bit mode"})
    torch_dtype: Optional[str] = field(default=None, metadata={"help": "The torch dtype for the pre-trained model"})
    device_map: str = field(default="auto", metadata={"help": "The device map for the pre-trained model"})

    # Method to process additional arguments
    def process_additional_args(self, addt_args: List[str]):
        """
        Process additional cmd line args and update the model configuration accordingly.
        """
        addt_args_dict = {}
        i = 0
        while i < len(addt_args):
            key = addt_args[i].lstrip('-')  # Remove leading dashes
            if i + 1 < len(addt_args) and not addt_args[i + 1].startswith('--'):
                value = addt_args[i + 1]
                i += 2  # Move past the current key-value pair
            else:
                value = True  # Assign a True value for standalone flags
                i += 1  # Move to the next item

            addt_args_dict[key] = value

        # Update the ModelConfig instance with the additional args
        self.__dict__.update(addt_args_dict)
    
    def __post_init__(self):
        """
        Post-initialization to validate some ModelConfig values
        """
        if self.torch_dtype and not hasattr(torch, self.torch_dtype):
            raise ValueError(f"Invalid torch dtype: {self.torch_dtype}")
        self.torch_dtype = getattr(torch, self.torch_dtype) if self.torch_dtype else None

@dataclass
class QuantizationConfig(BitsAndBytesConfig):
    quant_method: str = field(default="bitsandbytes", metadata={"help": "Quantization Method {bitsandbytes,gptq,awq}"})
    load_in_8bit: bool = field(default=False, metadata={"help": "Enable 8-bit quantization"})
    load_in_4bit: bool = field(default=False, metadata={"help": "Enable 4-bit quantization"})
    llm_int8_threshold: float = field(default=6.0, metadata={"help": "LLM.int8 threshold"})
    llm_int8_skip_modules: List[str] = field(default=None, metadata={"help": "Modules to skip for 8-bit conversion"})
    llm_int8_enable_fp32_cpu_offload: bool = field(default=False, metadata={"help": "Enable FP32 CPU offload for 8-bit"})
    llm_int8_has_fp16_weight: bool = field(default=False, metadata={"help": "Use FP16 weights for LLM.int8"})
    bnb_4bit_compute_dtype: str = field(default="float32", metadata={"help": "Compute dtype for 4-bit quantization"})
    bnb_4bit_quant_type: str = field(default="fp4", metadata={"help": "Quantization type for 4-bit"})
    bnb_4bit_use_double_quant: bool = field(default=False, metadata={"help": "Use double quantization for 4-bit"})

    def __post_init__(self):
        super().__init__(
            quant_method=self.quant_method,
            load_in_8bit=self.load_in_8bit,
            load_in_4bit=self.load_in_4bit,
            llm_int8_threshold=self.llm_int8_threshold,
            llm_int8_skip_modules=self.llm_int8_skip_modules,
            llm_int8_enable_fp32_cpu_offload=self.llm_int8_enable_fp32_cpu_offload,
            llm_int8_has_fp16_weight=self.llm_int8_has_fp16_weight,
            bnb_4bit_compute_dtype=self.bnb_4bit_compute_dtype,
            bnb_4bit_quant_type=self.bnb_4bit_quant_type,
            bnb_4bit_use_double_quant=self.bnb_4bit_use_double_quant,
        )

