# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import sys
from collections import defaultdict
import yaml
from dataclasses import asdict, fields
from transformers import HfArgumentParser, TrainingArguments
from cli import (DatasetConfig, ExtDataCollator, ExtLoraConfig, ModelConfig,
                 QuantizationConfig, TokenizerParams)

# Namespaces for each data class
NAMESPACES = {
    'MC': 'ModelConfig',
    'QC': 'QuantizationConfig',
    'ELC': 'ExtLoraConfig',
    'TA': 'TrainingArguments',
    'EDC': 'ExtDataCollator',
    'DC': 'DatasetConfig',
    'TP': 'TokenizerParams',
}

# Mapping from config section names to data classes
CONFIG_CLASS_MAP = {
    'ModelConfig': ModelConfig,
    'TokenizerParams': TokenizerParams,
    'QuantizationConfig': QuantizationConfig,
    'LoraConfig': ExtLoraConfig,
    'TrainingArguments': TrainingArguments,
    'DatasetConfig': DatasetConfig,
    'DataCollator': ExtDataCollator,
}

def filter_unsupported_init_args(dataclass_type, config_dict):
    supported_fields = {f.name for f in fields(dataclass_type) if f.init}
    filtered_config = {k: v for k, v in config_dict.items() if k in supported_fields}
    return filtered_config

# Function to extract and organize namespaced CLI arguments
def organize_cli_args(cli_args):
    organized_args = defaultdict(dict)
    for arg in cli_args:
        if arg.startswith('-'):
            key, value = arg.split('=')
            prefix, param = key.lstrip('-').split('_', 1)
            if prefix in NAMESPACES:
                class_name = NAMESPACES[prefix]
                organized_args[class_name][param] = value
    return organized_args


def parse_configs():
    # Capture raw CLI arguments (excluding the script name)
    raw_cli_args = sys.argv[1:]

    # Organize CLI arguments by their corresponding data classes
    organized_cli_args = organize_cli_args(raw_cli_args)

    # Parse and merge configurations
    parsed_configs = {}
    for section_name, class_type in CONFIG_CLASS_MAP.items():
        cli_args_for_section = {}
        if section_name in organized_cli_args:
            cli_args_for_section = organized_cli_args[section_name]

        filtered_config = filter_unsupported_init_args(CONFIG_CLASS_MAP[section_name], cli_args_for_section)
        parsed_configs[section_name] = CONFIG_CLASS_MAP[section_name](**filtered_config)

    return parsed_configs
