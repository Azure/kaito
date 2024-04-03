# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import os
from dataclasses import asdict
from datetime import datetime
import yaml

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

def flatten_config_to_cli_args(config, prefix=''):
    cli_args = []
    for key, value in config.items():
        if isinstance(value, dict):
            cli_args.extend(flatten_config_to_cli_args(value, prefix=f'{prefix}{key}_'))
        else:
            cli_arg = f'--{prefix}{key}'
            cli_args.append(cli_arg)
            cli_args.append(str(value))
    return cli_args

# Function to parse a single section
def parse_section(section_name, section_config, class_mapping):
    parser = HfArgumentParser((class_mapping[section_name],))
    # Flatten section_config to CLI-like arguments
    cli_args = flatten_config_to_cli_args(section_config, prefix='')
    # Parse the CLI-like arguments into a data class instance
    return parser.parse_args_into_dataclasses(cli_args)[0]

# Load the YAML configuration
CONFIG_YAML = os.environ.get('YAML_FILE_PATH', 'default_path_to_yaml')
with open(CONFIG_YAML, 'r') as file:
    full_config = yaml.safe_load(file)
training_config = full_config['training_config']

# Mapping from config section names to data classes
config_class_mapping = {
    'ModelConfig': ModelConfig,
    'TokenizerParams': TokenizerParams,
    'QuantizationConfig': QuantizationConfig,
    'LoraConfig': ExtLoraConfig,
    'TrainingArguments': TrainingArguments,
    'DatasetConfig': DatasetConfig,
    'DataCollator': ExtDataCollator,
}

# Parsing
parsed_configs = {}
for section_name, section_config in training_config.items():
    if section_name in config_class_mapping:
        parsed_configs[section_name] = parse_section(section_name, section_config, config_class_mapping)

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

# Loading the dataset
dataset = load_dataset(ds_config.dataset_name, split="train")

dataset = load("/data-volume")

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
