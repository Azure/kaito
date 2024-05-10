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

    def load_dataset(self):
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
        if self.dataset is None:
            raise ValueError("Dataset is not loaded.")
        self.dataset = self.dataset.shuffle(seed=seed)

    def split_dataset(self):
        if self.dataset is None:
            raise ValueError("Dataset is not loaded.")
        assert 0 < self.config.train_test_split <= 1, "Train/Test split needs to be between 0 and 1"
        if self.config.train_test_split < 1:
            split_dataset = self.dataset.train_test_split(
                test_size=1-self.config.train_test_split,
                seed=self.config.shuffle_seed
            )
            return split_dataset['train'], split_dataset['test']
        else:
            return self.dataset, None

    def get_dataset(self):
        if self.dataset is None:
            raise ValueError("Dataset is not loaded.")
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

    def format_text(self):
        if self.dataset is None:
            raise ValueError("Dataset is not loaded.")
        if self.config.response_column not in self.dataset.column_names:
            raise ValueError(f"Column '{self.config.response_column}' does not exist in the dataset. Available columns: {self.dataset.column_names}")
        self.dataset = self.dataset.select_columns([self.config.response_column])

    def format_instruct(self):
        """Ensure dataset is formatted for instruct fine tuning"""
        if self.dataset is None:
            raise ValueError("Dataset is not loaded.")        
        required_columns = [self.config.context_column, self.config.response_column]
        for column in required_columns:
            if column not in self.dataset.column_names:
                raise ValueError(f"Column '{column}' does not exist in the dataset. Available columns: {self.dataset.column_names}")
       
        # Select only the specified columns for fine tuning
        self.dataset = self.dataset.select_columns(required_columns)
       
        # Ensure correct column name
        if self.config.context_column != "prompt": 
            self.dataset = self.dataset.rename_column(self.config.context_column, "prompt")
        if self.config.response_column != "completion": 
            self.dataset = self.dataset.rename_column(self.config.context_column, "completion")
            

    def format_conversational(self):
        """Ensure some basic formatting of dataset for conversational fine tuning"""
        if self.dataset is None:
            raise ValueError("Dataset is not loaded.")
         
        # Check if the specified column exists in the dataset
        if self.config.messages_column not in self.dataset.column_names:
            raise ValueError(f"Column '{self.config.messages_column}' does not exist in the dataset. Available columns: {self.dataset.column_names}")
        
        # Select only the specified column for fine tuning
        self.dataset = self.dataset.select_columns([self.config.messages_column])
        
        # Ensure correct column name
        if self.config.messages_column != "messages": 
            self.dataset = self.dataset.rename_column(self.config.messages_column, "messages")

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
