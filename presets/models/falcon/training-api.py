# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import argparse
import os
from typing import List, Optional

import bitsandbytes as bnb
import torch
import transformers
import uvicorn
from datasets import load_dataset
from fastapi import FastAPI, HTTPException
from peft import (LoraConfig, PeftConfig, get_peft_model,
                  prepare_model_for_kbit_training)
from pydantic import BaseModel
from transformers import (AutoConfig, AutoModelForCausalLM, AutoTokenizer,
                          BitsAndBytesConfig)

parser = argparse.ArgumentParser(description='Falcon Model Configuration')
parser.add_argument('--load_in_8bit', default=False, action='store_true', help='Load model in 8-bit mode')
args = parser.parse_args()

app = FastAPI()

# Load the Pre-Trained Model
tokenizer = AutoTokenizer.from_pretrained("/workspace/tfs/weights")
tokenizer.pad_token = tokenizer.eos_token

bnb_config = BitsAndBytesConfig(
    load_in_4bit=True,
    load_4bit_use_double_quant=True,
    bnb_4bit_quant_type="nf4",
    bnb_4bit_compute_dtype=torch.bfloat16,
)

model = AutoModelForCausalLM.from_pretrained(
    "/workspace/tfs/weights", # args.model_id,
    device_map="auto",
    torch_dtype=torch.bfloat16,
    load_in_8bit=args.load_in_8bit,
    quantization_config=bnb_config,
)

# Preparing the Model for QLoRA
model = prepare_model_for_kbit_training(model)

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
model.config.use_cache = False
model.print_trainable_parameters()


# Loading and Preparing the Dataset 
def preprocess_data(example):
    prompt = f"<Human>: {example['Context']}\n<AI>: {example['Response']}".strip()
    return tokenizer(prompt, padding=True, truncation=True)

# def generate_prompt(data_point):
#   return f"""
# <Human>: {data_point["Context"]}
# <AI>: {data_point["Response"]}
#   """.strip()

# def generate_and_tokenize_prompt(data_point):
#   full_prompt = generate_prompt(data_point)
#   tokenized_full_prompt = tokenizer(full_prompt, padding=True, truncation=True)
#   return tokenized_full_prompt

dataset_name = 'Amod/mental_health_counseling_conversations'
dataset = load_dataset(dataset_name, split="train").shuffle()
dataset = dataset.map(preprocess_data)

# Setting Up the Training Arguments
training_args = transformers.TrainingArguments(
    # auto_find_batch_size=True, # Auto finds largest batch size that fits into memory
    per_device_train_batch_size=1,
    num_train_epochs=4, # Number of training epochs
    learning_rate=2e-4, # lr
    bf16=False, # precision
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
trainer.train()
trainer.save_model(".")