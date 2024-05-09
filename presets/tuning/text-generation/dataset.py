import json
import os
from typing import Optional

from datasets import load_dataset

SUPPORTED_EXTENSIONS = {'csv', 'json', 'parquet', 'arrow', 'webdataset'}

class DatasetManager:
    def __init__(self, config, tokenizer_params, preprocess_function: Optional[callable] = None):
        self.config = config
        self.tokenizer_params = tokenizer_params
        self.preprocess_function = preprocess_function
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

    def preprocess_data(self):
        if self.dataset is None:
            raise ValueError("Dataset is not loaded.")
        if self.preprocess_function:
            self.dataset = self.dataset.map(self.preprocess_function, batched=True, fn_kwargs=self.tokenizer_params)

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
    

    def format_text_fn(self, examples):
        # Use data from raw text column    
        return examples[self.config.response_column]

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

    def format_instruct_fn(self, examples): 
        output = []
        for i in range(len(examples[self.config.context_column])):
            prompt = examples[self.config.context_column][i]
            completion = examples[self.config.text_column][i]
            output.append({"prompt": prompt, "completion": completion})
        return output

    def format_conversational_fn(self, examples):
        output_data = []
        for item in examples[self.config.messages_column]:
            if isinstance(item, str):
                try:
                    # Attempt to parse the string as JSON
                    item = json.loads(item)
                except json.JSONDecodeError as e:
                    print(f"Skipping invalid JSON entry: {e}")
                    continue  # Skip the entry and move to the next
            
            # Check if the item is already properly structured
            if isinstance(item, dict) and 'messages' in item:
                json_data = item
            elif isinstance(item, list):
                # Wrap the list in a dict with a 'messages' key
                json_data = {'messages': item}
            else:
                print(f"Skipping entry with unsupported type or structure: {type(item)}")
                continue

            output_data.append(json_data)
        return output_data


