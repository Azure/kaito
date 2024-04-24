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


def filter_unsupported_init_args(dataclass_type, config_dict):
    supported_fields = {f.name for f in fields(dataclass_type) if f.init}
    filtered_config = {k: v for k, v in config_dict.items() if k in supported_fields}
    return filtered_config


# Function to parse a single section
def parse_section(section_name, section_config):
    parser = HfArgumentParser((CONFIG_CLASS_MAP[section_name],))
    # Flatten section_config to CLI-like arguments
    cli_args = flatten_config_to_cli_args(section_config, prefix='')
    # Parse the CLI-like arguments into a data class instance
    return parser.parse_args_into_dataclasses(cli_args)[0]


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


def merge_cli_args_with_yaml(cli_args, yaml_config):
    for key, value in cli_args.items():
        if key in yaml_config:
            # Value from CLI is a str needs to be converted
            old_value = yaml_config[key]
            if isinstance(old_value, bool):
                new_value = value.strip().lower() == "true"
            elif isinstance(old_value, int):
                new_value = int(value)
            elif isinstance(old_value, float):
                new_value = float(value)
            else:
                new_value = str(value)
            yaml_config[key] = new_value
    return yaml_config


def parse_configs(config_yaml):
    # Capture raw CLI arguments (excluding the script name)
    raw_cli_args = sys.argv[1:]
    print(raw_cli_args, "raw_cli_args")

    # Organize CLI arguments by their corresponding data classes
    organized_cli_args = organize_cli_args(raw_cli_args)
    print(organized_cli_args, "organized_cli_args")

    # Load the YAML configuration
    with open(config_yaml, 'r') as file:
        full_config = yaml.safe_load(file)
        print(full_config, "full_config")
    training_config = full_config.get('training_config', {})
    print(training_config, "training_config")

    # Parse and merge configurations
    parsed_configs = {}
    for section_name, class_type in CONFIG_CLASS_MAP.items():
        # Parse section from YAML
        yaml_parsed_instance = parse_section(section_name, training_config.get(section_name, {}))
        yaml_parsed_dict = asdict(yaml_parsed_instance)

        # Merge CLI args with YAML configs if CLI args are present
        if section_name in organized_cli_args:
            cli_args_for_section = organized_cli_args[section_name]
            merged_config = merge_cli_args_with_yaml(cli_args_for_section, yaml_parsed_dict)
        else:
            merged_config = yaml_parsed_dict

        filtered_config = filter_unsupported_init_args(CONFIG_CLASS_MAP[section_name], merged_config)
        parsed_configs[section_name] = CONFIG_CLASS_MAP[section_name](**filtered_config)

    return parsed_configs