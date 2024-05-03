# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import os
import sys
from dataclasses import asdict
from datetime import datetime

import torch
import transformers
from accelerate import Accelerator

from datasets import load_dataset
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
from transformers import (AutoModelForCausalLM, AutoTokenizer,
                          BitsAndBytesConfig, HfArgumentParser,
                          TrainingArguments)
from parser import parse_configs


DATASET_PATH = os.environ.get('DATASET_FOLDER_PATH', '/mnt/data')
CONFIG_YAML = os.environ.get('YAML_FILE_PATH', '/mnt/config/training_config.yaml')
parsed_configs = parse_configs(CONFIG_YAML)

model_config = parsed_configs.get('ModelConfig')
tk_params = parsed_configs.get('TokenizerParams')
bnb_config = parsed_configs.get('QuantizationConfig')
ext_lora_config = parsed_configs.get('LoraConfig')
ta_args = parsed_configs.get('TrainingArguments')
ds_config = parsed_configs.get('DatasetConfig')
dc_args = parsed_configs.get('DataCollator')

accelerator = Accelerator()

# Load Model Args
model_args = asdict(model_config)
if accelerator.distributed_type != "NO":  # Meaning we require distributed training
    print("Setting device map for distributed training")
    model_args["device_map"] = {"": Accelerator().process_index}

# Load BitsAndBytesConfig
bnb_config_args = asdict(bnb_config)
bnb_config = BitsAndBytesConfig(**bnb_config_args)
enable_qlora = bnb_config.is_quantizable()

# Load Tokenizer Params
tk_params = asdict(tk_params)

# Load the Pre-Trained Tokenizer
tokenizer = AutoTokenizer.from_pretrained(**model_args)
if not tokenizer.pad_token:
    tokenizer.pad_token = tokenizer.eos_token
if dc_args.mlm and tokenizer.mask_token is None:
    print(
        "This tokenizer does not have a mask token which is necessary for masked language modeling. "
        "You should pass `mlm=False` to train on causal language modeling instead. "
        "Setting mlm=False"
    )
    dc_args.mlm = False
dc_args.tokenizer = tokenizer

# Load the Pre-Trained Model
model = AutoModelForCausalLM.from_pretrained(
    **model_args,
    quantization_config=bnb_config if enable_qlora else None,
)

print("Model Loaded")

if enable_qlora:
    # Preparing the Model for QLoRA
    model = prepare_model_for_kbit_training(model)
    print("QLoRA Enabled")

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

found_dataset = False
# Loading the dataset
if ds_config.dataset_path:
    DATASET_PATH = os.path.join("/mnt", ds_config.dataset_path.strip("/"))
    found_dataset = True
else:
    # Find a valid file
    VALID_EXTENSIONS = ('.csv', '.json', '.txt', '.parquet', '.xls', '.xlsx', '.arrow')
    default_path = DATASET_PATH
    for root, dirs, files in os.walk(default_path):
        for file in files:
            if file.endswith(VALID_EXTENSIONS):
                DATASET_PATH = os.path.join(root, file)
                found_dataset = True
                break

# Check if DATASET_PATH has been set successfully
if found_dataset:
    print(f"Dataset found: {DATASET_PATH}")
else:
    print("No valid dataset file found.")
    raise ValueError(f"Unable to find a valid dataset file.")

if ds_config.dataset_extension:
    file_ext = ds_config.dataset_extension
else:
    _, file_ext = os.path.splitext(DATASET_PATH)
    file_ext = file_ext[1:] if file_ext else file_ext # Remove leading "." from file ext name
dataset = load_dataset(file_ext, data_files=DATASET_PATH, split="train")

# Shuffling the dataset (if needed)
if ds_config.shuffle_dataset:
    dataset = dataset.shuffle(seed=ds_config.shuffle_seed)

# Preprocessing the data
dataset = dataset.map(preprocess_data)

assert 0 < ds_config.train_test_split <= 1, "Train/Test split needs to be between 0 and 1"

# Initialize variables for train and eval datasets
train_dataset, eval_dataset = dataset, None

if ds_config.train_test_split < 1:
    # Splitting the dataset into training and test sets
    split_dataset = dataset.train_test_split(
        test_size=1-ds_config.train_test_split,
        seed=ds_config.shuffle_seed
    )
    train_dataset, eval_dataset = split_dataset['train'], split_dataset['test']
    print("Training Dataset Dimensions: ", train_dataset.shape)
    print("Test Dataset Dimensions: ", eval_dataset.shape)
else:
    print(f"Using full dataset for training. Dimensions: {train_dataset.shape}")

# checkpoint_callback = CheckpointCallback()

# Training the Model
trainer = accelerator.prepare(transformers.Trainer(
    model=model,
    train_dataset=train_dataset,
    eval_dataset=eval_dataset,
    args=ta_args,
    data_collator=dc_args,
    # callbacks=[checkpoint_callback]
))
trainer.train()
os.makedirs(ta_args.output_dir, exist_ok=True)
trainer.save_model(ta_args.output_dir)

# Write file to signify training completion
timestamp = datetime.now().strftime("%Y-%m-%d-%H-%M-%S")
completion_indicator_path = os.path.join(ta_args.output_dir, "fine_tuning_completed.txt")
with open(completion_indicator_path, 'w') as f:
    f.write(f"Fine-Tuning completed at {timestamp}\n")
