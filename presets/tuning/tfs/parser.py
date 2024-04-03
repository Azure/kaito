# parser.py
import sys
from collections import defaultdict
import yaml
from dataclasses import asdict
from transformers import HfArgumentParser, TrainingArguments
from cli import (DatasetConfig, ExtDataCollator, ExtLoraConfig, ModelConfig,
                 QuantizationConfig, TokenizerParams)

# Namespaces for each data class
namespaces = {
    'MC': 'ModelConfig',
    'QC': 'QuantizationConfig',
    'ELC': 'ExtLoraConfig',
    'TA': 'TrainingArguments',
    'EDC': 'ExtDataCollator',
    'DC': 'DatasetConfig',
    'TP': 'TokenizerParams',
}

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


# Function to extract and organize namespaced CLI arguments
def organize_cli_args(cli_args, namespaces):
    organized_args = defaultdict(dict)
    for arg in cli_args:
        if arg.startswith('-'):
            key, value = arg.split('=')
            prefix, param = key.lstrip('-').split('_', 1)
            if prefix in namespaces:
                class_name = namespaces[prefix]
                organized_args[class_name][param] = value
    return organized_args

def merge_cli_args_with_yaml(cli_args, yaml_config):
    for key, value in cli_args.items():
        # Type conversions can be added here based on the expected type in yaml_config
        yaml_config[key] = value
    return yaml_config

def parse_configs(CONFIG_YAML):
    # Capture raw CLI arguments (excluding the script name)
    raw_cli_args = sys.argv[1:]

    # Organize CLI arguments by their corresponding data classes
    organized_cli_args = organize_cli_args(raw_cli_args, namespaces)

    # Load the YAML configuration
    with open(CONFIG_YAML, 'r') as file:
        full_config = yaml.safe_load(file)
    training_config = full_config.get('training_config', {})

    # Parse and merge configurations
    parsed_configs = {}
    for section_name, class_type in config_class_mapping.items():
        # Parse section from YAML
        yaml_parsed_instance = parse_section(section_name, training_config.get(section_name, {}), config_class_mapping)
        yaml_parsed_dict = asdict(yaml_parsed_instance)

        # Merge CLI args with YAML configs if CLI args are present
        if section_name in organized_cli_args:
            cli_args_for_section = organized_cli_args[section_name]
            merged_config = merge_cli_args_with_yaml(cli_args_for_section, yaml_parsed_dict)
        else:
            merged_config = yaml_parsed_dict

        # Create a final instance of the data class with the merged configuration
        parsed_configs[section_name] = config_class_mapping[section_name](**merged_config)

    return parsed_configs
