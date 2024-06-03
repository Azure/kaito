# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import os
import sys
from dataclasses import asdict
from datetime import datetime
from parser import parse_configs

import torch
import transformers
from accelerate import Accelerator
from dataset import DatasetManager
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
from transformers import (AutoModelForCausalLM, AutoTokenizer,
                          BitsAndBytesConfig, HfArgumentParser, Trainer,
                          TrainingArguments)
from trl import SFTTrainer

CONFIG_YAML = os.environ.get('YAML_FILE_PATH', '/mnt/config/training_config.yaml')

parsed_configs = parse_configs(CONFIG_YAML)

model_config = parsed_configs.get('ModelConfig')
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
    model_args["device_map"] = {"": accelerator.process_index}

# Load BitsAndBytesConfig
bnb_config_args = asdict(bnb_config)
bnb_config = BitsAndBytesConfig(**bnb_config_args)
enable_qlora = bnb_config.is_quantizable()

# Load the Pre-Trained Tokenizer
tokenizer_args = {key: value for key, value in model_args.items() if key != "torch_dtype"}
tokenizer = AutoTokenizer.from_pretrained(**tokenizer_args)
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

if not ext_lora_config:
    raise ValueError("LoraConfig must be specified")

lora_config_args = asdict(ext_lora_config)
lora_config = LoraConfig(**lora_config_args)

model = get_peft_model(model, lora_config)
# Cache is only used for generation, not for training
model.config.use_cache = False
model.print_trainable_parameters()

dm = DatasetManager(ds_config, tokenizer)
# Load the dataset
dm.load_data()
if not dm.get_dataset():
    print("Failed to load dataset.")
    raise ValueError("Unable to load the dataset.")

# Shuffling the dataset (if needed)
if ds_config.shuffle_dataset:
    dm.shuffle_dataset()

dm.format_and_preprocess()
train_dataset, eval_dataset = dm.split_dataset()

# checkpoint_callback = CheckpointCallback()

# Prepare for training
torch.cuda.set_device(accelerator.process_index)
torch.cuda.empty_cache()
# Training the Model
trainer = accelerator.prepare(SFTTrainer(
    model=model,
    tokenizer=tokenizer,
    train_dataset=train_dataset,
    eval_dataset=eval_dataset,
    args=ta_args,
    data_collator=dc_args,
    dataset_text_field=dm.dataset_text_field,
    # metrics = "tensorboard" or "wandb" # TODO
))
trainer.train()
os.makedirs(ta_args.output_dir, exist_ok=True)
trainer.save_model(ta_args.output_dir)

# Write file to signify training completion
timestamp = datetime.now().strftime("%Y-%m-%d-%H-%M-%S")
completion_indicator_path = os.path.join(ta_args.output_dir, "fine_tuning_completed.txt")
with open(completion_indicator_path, 'w') as f:
    f.write(f"Fine-Tuning completed at {timestamp}\n")
