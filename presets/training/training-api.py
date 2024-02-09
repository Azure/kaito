# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import argparse
import os
from dataclasses import asdict, dataclass, field
from typing import Any, Dict, List, Optional

import bitsandbytes as bnb
import torch
import transformers
import uvicorn
from accelerate import Accelerator
from cli import ExtDataCollator, ExtLoraConfig, ModelConfig, QuantizationConfig
from datasets import load_dataset
from fastapi import FastAPI, HTTPException
from peft import (LoraConfig, PeftConfig, get_peft_model,
                  prepare_model_for_kbit_training)
from pydantic import BaseModel
from transformers import (AutoConfig, AutoModelForCausalLM, AutoTokenizer,
                          BitsAndBytesConfig, DataCollatorForLanguageModeling,
                          HfArgumentParser, TrainingArguments)

parser = HfArgumentParser((ModelConfig, QuantizationConfig, ExtLoraConfig, TrainingArguments, ExtDataCollator))
# TODO: How to pass dict/lists/other ds on cmd line
model_config, bnb_config, ext_lora_config, ta_args, dc_args, additional_args = parser.parse_args_into_dataclasses(
    return_remaining_strings=True
)

print("Additional arguments:", additional_args)

# model_config.process_additional_args(additional_args)
model_args = asdict(model_config)
model_args["local_files_only"] = not model_args.pop('allow_remote_files')
model_args["revision"] = model_args.pop('m_revision')
model_args["load_in_4bit"] = model_args.pop('m_load_in_4bit')
model_args["load_in_8bit"] = model_args.pop('m_load_in_8bit')


accelerator = Accelerator()

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

# TODO: How to abstract this into API
# Loading and Preparing the Dataset 
def preprocess_data(example):
    prompt = f"<human>: {example['Context']}\n<assistance>: {example['Response']}".strip()
    return tokenizer(prompt, padding=True, truncation=True)

dataset_name = 'Amod/mental_health_counseling_conversations'
dataset = load_dataset(dataset_name, split="train")
# Shuffle and preprocess the data
dataset = dataset.shuffle().map(preprocess_data)
print("Dataset Dimensions: ", dataset.shape)

# Setting Up the Training Arguments
# training_args = transformers.TrainingArguments(
#     # auto_find_batch_size=True, # Auto finds largest batch size that fits into memory
#     # gradient_checkpointing_kwargs={"use_reentrant": False},
#     # gradient_checkpointing=False,
#     # ddp_backend="nccl",
#     # ddp_find_unused_parameters=False,
#     # per_device_train_batch_size=1,
#     num_train_epochs=4, # Number of training epochs
#     learning_rate=2e-4, # lr
#     bf16=False, # precision
#     save_total_limit=4, # Total # of ckpts to save
#     logging_steps=10, # # of steps between each logging
#     output_dir=".", #  Dir where model ckpts saved
#     save_strategy='epoch', # Strategy for saving ckpts. Here ckpt saved after each epoch
# )

# print("1", training_args)

# Needs to be explicitly defined
# if not targs.gradient_checkpointing_kwargs:
#     from functools import partial
#     from torch.utils import checkpoint
#     notfailing_checkpoint = partial(checkpoint.checkpoint, use_reentrant=False)
#     checkpoint.checkpoint = notfailing_checkpoint


print("2", ta_args)

# Training the Model
trainer = transformers.Trainer(
    model=model,
    train_dataset=dataset,
    args=ta_args,
    data_collator=dc_args,
)
trainer.train()
trainer.save_model(".")
