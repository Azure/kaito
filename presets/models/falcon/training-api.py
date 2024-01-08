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

