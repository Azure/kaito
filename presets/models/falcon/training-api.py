# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import argparse
import os
from typing import List, Optional

import torch
import transformers
import uvicorn
from datasets import load_dataset
from fastapi import FastAPI, HTTPException
from peft import (LoraConfig, PeftConfig, get_peft_model,
                  prepare_model_for_kbit_training)
from pydantic import BaseModel
from transformers import AutoConfig, AutoModelForCausalLM, AutoTokenizer

parser = argparse.ArgumentParser(description='Falcon Model Configuration')
parser.add_argument('--load_in_8bit', default=False, action='store_true', help='Load model in 8-bit mode')
# parser.add_argument('--model_id', required=True, type=str, help='The Falcon ID for the pre-trained model')
args = parser.parse_args()

app = FastAPI()

# Load the Pre-Trained Model
tokenizer = AutoTokenizer.from_pretrained("/workspace/tfs/weights")
model = AutoModelForCausalLM.from_pretrained(
    "/workspace/tfs/weights", # args.model_id,
    device_map="auto",
    torch_dtype=torch.bfloat16,
    load_in_8bit=args.load_in_8bit,
)

# Configuring LoRA
config = LoraConfig(
    r=16,
    lora_alpha=32,
    target_modules=["query_key_value"],
    lora_dropout=0.05,
    bias="none",
    task_type="CAUSAL_LM"
)

model = get_peft_model(model, config)
model.print_trainable_parameters()


# Loading and Preparing the Dataset 

def generate_prompt(data_point):
  return f"""
<Human>: {data_point["Context"]}
<AI>: {data_point["Response"]}
  """.strip()

def generate_and_tokenize_prompt(data_point):
  full_prompt = generate_prompt(data_point)
  tokenized_full_prompt = tokenizer(full_prompt, padding=True, truncation=True)
  return tokenized_full_prompt

from datasets import load_dataset

dataset_name = 'Amod/mental_health_counseling_conversations'
dataset = load_dataset(dataset_name, split="train")

dataset = dataset.shuffle().map(generate_and_tokenize_prompt)

# Setting Up the Training Arguments
training_args = transformers.TrainingArguments(
    auto_find_batch_size=True, # Auto finds largest batch size that fits into memory
    num_train_epochs=4, # # of training epichs
    learning_rate=2e-4, # lr
    bf16=True, # precision
    save_total_limit=4, # Total # of ckpts to save
    logging_steps=10, # # of steps between each logging
    output_dir=".", #  Dir where model ckpts saved
    save_strategy='epoch', # Strategy for saving ckpts. Here ckpt saved after each epoch
)

# Training the Model
trainer = transformers.Trainer(
    model=model,
    train_dataset=dataset,
    args=training_args,
    data_collator=transformers.DataCollatorForLanguageModeling(tokenizer, mlm=False),
)
model.config.use_cache = False
trainer.train()