# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
from dataclasses import asdict

import torch
import transformers
from accelerate import Accelerator
from cli import (DatasetConfig, ExtDataCollator, ExtLoraConfig, ModelConfig,
                 QuantizationConfig, TokenizerParams)
from datasets import load_dataset
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
from transformers import (AutoModelForCausalLM, AutoTokenizer,
                          BitsAndBytesConfig, HfArgumentParser,
                          TrainingArguments)

# Parsing
parser = HfArgumentParser((ModelConfig, QuantizationConfig, ExtLoraConfig, TrainingArguments, ExtDataCollator, DatasetConfig, TokenizerParams))
# TODO: How to pass dict/lists/other ds on cmd line
model_config, bnb_config, ext_lora_config, ta_args, dc_args, ds_config, tk_params, additional_args = parser.parse_args_into_dataclasses(
    return_remaining_strings=True
)

print("Unmatched arguments:", additional_args)

accelerator = Accelerator()

# Load Model Args
model_args = asdict(model_config)
model_args["local_files_only"] = not model_args.pop('allow_remote_files')
model_args["revision"] = model_args.pop('m_revision')
model_args["load_in_4bit"] = model_args.pop('m_load_in_4bit')
model_args["load_in_8bit"] = model_args.pop('m_load_in_8bit')

# Load BitsAndBytesConfig
bnb_config_args = asdict(bnb_config)
bnb_config = BitsAndBytesConfig(**bnb_config_args)
enable_qlora = bnb_config.is_quantizable()

# Load the Pre-Trained Tokenizer
tokenizer = AutoTokenizer.from_pretrained(**model_args)
if not tokenizer.pad_token:
    tokenizer.pad_token = tokenizer.eos_token
dc_args.tokenizer = tokenizer

# Load the Pre-Trained Model
model = AutoModelForCausalLM.from_pretrained(
    **model_args,
    quantization_config=bnb_config if enable_qlora else None,
)

# Load Tokenizer Params
tk_params = asdict(tk_params)
tk_params["pad_to_multiple_of"] = tk_params.pop("tok_pad_to_multiple_of")
tk_params["return_tensors"] = tk_params.pop("tok_return_tensors")

if enable_qlora:
    # Preparing the Model for QLoRA
    model = prepare_model_for_kbit_training(model)

assert ext_lora_config is not None, "LoraConfig must be specified"
lora_config_args = asdict(ext_lora_config)
lora_config = LoraConfig(**lora_config_args)

model = get_peft_model(model, lora_config)
# Cache is only used for generation, not for training
model.config.use_cache = False
model.print_trainable_parameters()

# Loading and Preparing the Dataset 
# Data format: https://huggingface.co/docs/autotrain/en/llm_finetuning
def preprocess_data(example):
    prompt = f"human: {example[ds_config.context_column]}\n bot: {example[ds_config.response_column]}".strip()
    return tokenizer(prompt, **tk_params)

# Loading the dataset
dataset = load_dataset(ds_config.dataset_name, split="train")

# Shuffling the dataset (if needed)
if ds_config.shuffle_dataset: 
    dataset = dataset.shuffle(seed=ds_config.shuffle_seed)

# Preprocessing the data
dataset = dataset.map(preprocess_data)

assert 0 <= ds_config.train_test_split <= 1, "Train/Test split needs to be between 0 and 1"
# Splitting the dataset into training and test sets
split_dataset = dataset.train_test_split(
    test_size=1-ds_config.train_test_split, 
    seed=ds_config.shuffle_seed
)

print("Training Dataset Dimensions: ", split_dataset['train'].shape)
print("Test Dataset Dimensions: ", split_dataset['test'].shape)
print("Example Dataset Entry: ", split_dataset['train'][0])

print("Training Arguments: ", ta_args)

# Training the Model
trainer = transformers.Trainer(
    model=model,
    train_dataset=split_dataset['train'],
    eval_dataset=split_dataset['test'],
    args=ta_args,
    data_collator=dc_args,
)
trainer.train()
trainer.save_model(".")
