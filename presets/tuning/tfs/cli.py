# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import os
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Dict, List, Optional

import torch
from peft import LoraConfig
from transformers import (BitsAndBytesConfig, DataCollatorForLanguageModeling,
                          PreTrainedTokenizer, TrainerCallback)


@dataclass
class ExtDataCollator(DataCollatorForLanguageModeling):
    tokenizer: Optional[PreTrainedTokenizer] = field(default=PreTrainedTokenizer, metadata={"help": "Tokenizer for DataCollatorForLanguageModeling"})

@dataclass
class ExtLoraConfig(LoraConfig):
    """
    Lora Config
    """
    init_lora_weights: bool = field(default=True, metadata={"help": "Enable initialization of LoRA weights"})
    target_modules: Optional[List[str]] = field(default=None, metadata={"help": ("List of module names to replace with LoRA.")})
    layers_to_transform: Optional[List[int]] = field(default=None, metadata={"help": "Layer indices to apply LoRA"})
    layers_pattern: Optional[List[str]] = field(default=None, metadata={"help": "Pattern to match layers for LoRA"})
    loftq_config: Dict[str, any] = field(default_factory=dict, metadata={"help": "LoftQ configuration for quantization"})

@dataclass
class DatasetConfig: 
    """
    Config for Dataset 
    """
    dataset_name: str = field(metadata={"help": "Name of Dataset"})
    shuffle_dataset: bool = field(default=True, metadata={"help": "Whether to shuffle dataset"})
    shuffle_seed: int = field(default=42, metadata={"help": "Seed for shuffling data"})
    context_column: str = field(default="Context", metadata={"help": "Example human input column in the dataset"})
    response_column: str = field(default="Response", metadata={"help": "Example bot response output column in the dataset"})
    train_test_split: float = field(default=0.8, metadata={"help": "Split between test and training data (e.g. 0.8 means 80/20% train/test split)"})

@dataclass
class TokenizerParams:
    """
    Tokenizer params 
    """
    add_special_tokens: bool = field(default=True, metadata={"help": ""})
    padding: bool = field(default=False, metadata={"help": ""})
    truncation: bool = field(default=None, metadata={"help": ""})
    max_length: Optional[int] = field(default=None, metadata={"help": ""})
    stride: int = field(default=0, metadata={"help": ""})
    is_split_into_words: bool = field(default=False, metadata={"help": ""})
    pad_to_multiple_of: Optional[int] = field(default=None, metadata={"help": ""})
    return_tensors: Optional[str] = field(default=None, metadata={"help": ""})
    return_token_type_ids: Optional[bool] = field(default=None, metadata={"help": ""})
    return_attention_mask: Optional[bool] = field(default=None, metadata={"help": ""})
    return_overflowing_tokens: bool = field(default=False, metadata={"help": ""})
    return_special_tokens_mask: bool = field(default=False, metadata={"help": ""})
    return_offsets_mapping: bool = field(default=False, metadata={"help": ""})
    return_length: bool = field(default=False, metadata={"help": ""})
    verbose: bool = field(default=True, metadata={"help": ""})

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
    local_files_only: bool = field(default=False, metadata={"help": "Allow using remote files, default is local only"})
    revision: str = field(default="main", metadata={"help": "Specific model version to use"})
    trust_remote_code: bool = field(default=False, metadata={"help": "Enable trusting remote code when loading the model"})
    load_in_4bit: bool = field(default=False, metadata={"help": "Load model in 4-bit mode"})
    load_in_8bit: bool = field(default=False, metadata={"help": "Load model in 8-bit mode"})
    torch_dtype: Optional[str] = field(default=None, metadata={"help": "The torch dtype for the pre-trained model"})
    device_map: str = field(default="auto", metadata={"help": "The device map for the pre-trained model"})

    def __post_init__(self):
        """
        Post-initialization to validate some ModelConfig values
        """
        if self.torch_dtype and not hasattr(torch, self.torch_dtype):
            raise ValueError(f"Invalid torch dtype: {self.torch_dtype}")
        self.torch_dtype = getattr(torch, self.torch_dtype) if self.torch_dtype else None

@dataclass
class QuantizationConfig(BitsAndBytesConfig):
    """
    Quanitization Configuration
    """
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

# class CheckpointCallback(TrainerCallback):
#     def on_train_end(self, args, state, control, **kwargs):
#         model_path = args.output_dir
#         timestamp = datetime.now().strftime("%Y-%m-%d-%H-%M-%S")
#         img_tag = f"ghcr.io/YOUR_USERNAME/LoRA-Adapter:{timestamp}"
        
#         # Write a file to indicate fine_tuning completion
#         completion_indicator_path = os.path.join(model_path, "training_completed.txt")
#         with open(completion_indicator_path, 'w') as f:
#             f.write(f"Training completed at {timestamp}\n")
#             f.write(f"Image Tag: {img_tag}\n")

    # This method is called whenever a checkpoint is saved.
    # def on_save(self, args, state, control, **kwargs):
    #     docker_build_and_push()