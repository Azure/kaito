# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import os
from typing import Optional

from datasets import load_dataset

SUPPORTED_EXTENSIONS = {'csv', 'json', 'parquet', 'arrow', 'webdataset'}

class DatasetManager:
    def __init__(self, config, tokenizer, tokenizer_params):
        self.config = config
        self.tokenizer_params = tokenizer_params
        self.tokenizer = tokenizer
        self.dataset = None
        self.dataset_text_field = None # Set this field if dataset consists of singular text column

    def check_dataset_loaded(self):
        if self.dataset is None:
            raise ValueError("Dataset is not loaded.")

    def check_column_exists(self, column_name):
        if column_name not in self.dataset.column_names:
            raise ValueError(f"Column '{column_name}' does not exist in the dataset. Available columns: {self.dataset.column_names}")

    def select_and_rename_columns(self, columns_to_select, rename_map=None):
        self.dataset = self.dataset.select_columns(columns_to_select)
        if rename_map:
            for old_name, new_name in rename_map.items():
                if old_name != new_name:
                    self.dataset = self.dataset.rename_column(old_name, new_name)

    def load_data(self):
        if self.config.dataset_path:
            dataset_path = os.path.join("/mnt", self.config.dataset_path.strip("/"))
        else:
            dataset_path = self.find_valid_dataset(os.environ.get('DATASET_FOLDER_PATH', '/mnt/data'))
            if not dataset_path:
                raise ValueError("Unable to find a valid dataset file.")

        file_ext = self.config.dataset_extension if self.config.dataset_extension else self.get_file_extension(dataset_path)
        try:
            self.dataset = load_dataset(file_ext, data_files=dataset_path, split="train")
            print(f"Dataset loaded successfully from {dataset_path} with file type '{file_ext}'.")
        except Exception as e:
            print(f"Error loading dataset: {e}")
            raise ValueError(f"Unable to load dataset {dataset_path} with file type '{file_ext}'")

    def find_valid_dataset(self, data_dir):
        """ Searches for a file with a valid dataset type in the given directory. """
        for root, dirs, files in os.walk(data_dir):
            for file in files:
                filename_lower = file.lower()  # Convert to lowercase once per filename
                for ext in SUPPORTED_EXTENSIONS:
                    if ext in filename_lower:
                        return os.path.join(root, file)
        return None

    def get_file_extension(self, file_path):
        """ Returns the file extension based on filetype guess or filename. """
        filename_lower = os.path.basename(file_path).lower()
        for ext in SUPPORTED_EXTENSIONS:
            if ext in filename_lower:
                return ext
        _, file_ext = os.path.splitext(file_path)
        return file_ext[1:]  # Remove leading "."

    def shuffle_dataset(self, seed=None):
        self.check_dataset_loaded()
        self.dataset = self.dataset.shuffle(seed=seed)

    def split_dataset(self):
        self.check_dataset_loaded()
        if not (0 < self.config.train_test_split <= 1):
            raise ValueError("Train/Test split needs to be between 0 and 1")
        if self.config.train_test_split < 1:
            split_dataset = self.dataset.train_test_split(
                test_size=1-self.config.train_test_split,
                seed=self.config.shuffle_seed
            )
            return split_dataset['train'], split_dataset['test']
        else:
            return self.dataset, None

    def get_dataset(self):
        self.check_dataset_loaded()
        return self.dataset

    def format_and_preprocess(self):
        # OAI Compliant: https://platform.openai.com/docs/guides/fine-tuning/preparing-your-dataset
        # https://github.com/huggingface/trl/blob/main/trl/extras/dataset_formatting.py
        # https://huggingface.co/docs/trl/en/sft_trainer#dataset-format-support
        if self.config.messages_column:
            self.format_conversational()
        elif self.config.context_column and self.config.response_column:
            self.format_instruct()
        elif self.config.response_column:
            self.dataset = self.dataset.map(
                lambda example: self.tokenizer(example[self.config.response_column], **self.tokenizer_params),
                batched=True
            )
            self.format_text()
            self.dataset_text_field = self.config.response_column

    def format_text(self):
        self.check_dataset_loaded()
        self.check_column_exists(self.config.response_column)
        self.select_and_rename_columns([self.config.response_column])

    def format_instruct(self):
        """Ensure dataset is formatted for instruct fine tuning"""
        self.check_dataset_loaded()
        required_columns = [self.config.context_column, self.config.response_column]
        for column in required_columns:
            self.check_column_exists(column)

        # Select and rename columns
        rename_map = {}
        if self.config.context_column != "prompt":
            rename_map[self.config.context_column] = "prompt"
        if self.config.response_column != "completion":
            rename_map[self.config.response_column] = "completion"
        self.select_and_rename_columns(required_columns, rename_map)

    def format_conversational(self):
        """Ensure some basic formatting of dataset for conversational fine tuning"""
        self.check_dataset_loaded()
        # Check if the specified column exists in the dataset
        self.check_column_exists(self.config.messages_column)

         # Select and rename columns
        rename_map = {}
        if self.config.messages_column != "messages":
            rename_map[self.config.messages_column] = "messages"
        self.select_and_rename_columns([self.config.messages_column], rename_map)

    # Consider supporting in future
    # https://github.com/huggingface/trl/pull/444
    # def format_instruction_based_fn(self, examples): 
    #     output_text = []
    #     for i in range(len(examples[self.config.context_column])):
    #         instruction = examples[self.config.instruction_column][i]
    #         context = examples[self.config.context_column][i]
    #         response = examples[self.config.response_column][i]
    #         text = f'''Below is an instruction that describes a task, paired with an input that provides further context. Write a response that appropriately completes the request.

    #         ### Instruction:
    #         {instruction}
            
    #         ### Input:
    #         {context}
            
    #         ### Response:
    #         {response}
    #         '''
    #         output_text.append(text)
    #     return output_text
