# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import os
from dataclasses import asdict, dataclass, field
from typing import Any, Dict, List, Optional

import GPUtil
import torch
import transformers
import uvicorn
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Extra, Field
from transformers import (AutoModelForCausalLM, AutoTokenizer,
                          GenerationConfig, HfArgumentParser)


@dataclass
class ModelConfig:
    """
    Transformers Model Configuration Parameters
    """
    pipeline: str = field(metadata={"help": "The model pipeline for the pre-trained model"})
    pretrained_model_name_or_path: Optional[str] = field(default="/workspace/tfs/weights", metadata={"help": "Path to the pretrained model or model identifier from huggingface.co/models"})
    state_dict: Optional[Dict[str, Any]] = field(default=None, metadata={"help": "State dictionary for the model"})
    cache_dir: Optional[str] = field(default=None, metadata={"help": "Cache directory for the model"})
    from_tf: bool = field(default=False, metadata={"help": "Load model from a TensorFlow checkpoint"})
    force_download: bool = field(default=False, metadata={"help": "Force the download of the model"})
    resume_download: bool = field(default=False, metadata={"help": "Resume an interrupted download"})
    proxies: Optional[str] = field(default=None, metadata={"help": "Proxy configuration for downloading the model"})
    output_loading_info: bool = field(default=False, metadata={"help": "Output additional loading information"})
    allow_remote_files: bool = field(default=False, metadata={"help": "Allow using remote files, default is local only"})
    revision: str = field(default="main", metadata={"help": "Specific model version to use"})
    trust_remote_code: bool = field(default=False, metadata={"help": "Enable trusting remote code when loading the model"})
    load_in_4bit: bool = field(default=False, metadata={"help": "Load model in 4-bit mode"})
    load_in_8bit: bool = field(default=False, metadata={"help": "Load model in 8-bit mode"})
    torch_dtype: Optional[str] = field(default=None, metadata={"help": "The torch dtype for the pre-trained model"})
    device_map: str = field(default="auto", metadata={"help": "The device map for the pre-trained model"})
    
    # Method to process additional arguments
    def process_additional_args(self, addt_args: List[str]):
        """
        Process additional cmd line args and update the model configuration accordingly.
        """
        addt_args_dict = {}
        i = 0
        while i < len(addt_args):
            key = addt_args[i].lstrip('-')  # Remove leading dashes
            if i + 1 < len(addt_args) and not addt_args[i + 1].startswith('--'):
                value = addt_args[i + 1]
                i += 2  # Move past the current key-value pair
            else:
                value = True  # Assign a True value for standalone flags
                i += 1  # Move to the next item
            
            addt_args_dict[key] = value

        # Update the ModelConfig instance with the additional args
        self.__dict__.update(addt_args_dict)

    def __post_init__(self):
        """
        Post-initialization to validate some ModelConfig values
        """
        if self.torch_dtype and not hasattr(torch, self.torch_dtype):
            raise ValueError(f"Invalid torch dtype: {self.torch_dtype}")
        self.torch_dtype = getattr(torch, self.torch_dtype) if self.torch_dtype else None

        supported_pipelines = {"conversational", "text-generation"}
        if self.pipeline not in supported_pipelines:
            raise ValueError(f"Unsupported pipeline: {self.pipeline}")

parser = HfArgumentParser(ModelConfig)
args, additional_args = parser.parse_args_into_dataclasses(
    return_remaining_strings=True
)

args.process_additional_args(additional_args)

model_args = asdict(args)
model_args["local_files_only"] = not model_args.pop('allow_remote_files')
model_pipeline = model_args.pop('pipeline')

app = FastAPI()
tokenizer = AutoTokenizer.from_pretrained(**model_args)
model = AutoModelForCausalLM.from_pretrained(**model_args)

pipeline_kwargs = {
    "trust_remote_code": args.trust_remote_code,
    "device_map": args.device_map,
}

if args.torch_dtype:
    pipeline_kwargs["torch_dtype"] = args.torch_dtype

pipeline = transformers.pipeline(
    model_pipeline,
    model=model,
    tokenizer=tokenizer,
    **pipeline_kwargs
)

try:
    # Attempt to load the generation configuration
    default_generate_config = GenerationConfig.from_pretrained(
        args.pretrained_model_name_or_path, 
        local_files_only=args.local_files_only
    ).to_dict()
except Exception as e:
    default_generate_config = {}

@app.get('/')
def home():
    return "Server is running", 200

@app.get("/healthz")
def health_check():
    if not torch.cuda.is_available():
        raise HTTPException(status_code=500, detail="No GPU available")
    if not model:
        raise HTTPException(status_code=500, detail="Model not initialized")
    if not pipeline:
        raise HTTPException(status_code=500, detail="Pipeline not initialized")
    return {"status": "Healthy"}

class GenerateKwargs(BaseModel):
    max_length: int = 200
    min_length: int = 0
    do_sample: bool = True
    early_stopping: bool = False
    num_beams: int = 1
    num_beam_groups: int = 1
    diversity_penalty: float = 0.0
    temperature: float = 1.0
    top_k: int = 10
    top_p: float = 1
    typical_p: float = 1
    repetition_penalty: float = 1
    length_penalty: float = 1
    no_repeat_ngram_size: int = 0
    encoder_no_repeat_ngram_size: int = 0
    bad_words_ids: Optional[List[int]] = None
    num_return_sequences: int = 1
    output_scores: bool = False
    return_dict_in_generate: bool = False
    pad_token_id: Optional[int] = tokenizer.pad_token_id
    eos_token_id: Optional[int] = tokenizer.eos_token_id
    forced_bos_token_id: Optional[int] = None
    forced_eos_token_id: Optional[int] = None
    remove_invalid_values: Optional[bool] = None
    class Config:
        extra = Extra.allow # Allows for additional fields not explicitly defined

class UnifiedRequestModel(BaseModel):
    # Fields for text generation
    prompt: Optional[str] = Field(None, description="Prompt for text generation")
    return_full_text: Optional[bool] = Field(True, description="Return full text if True, else only added text")
    clean_up_tokenization_spaces: Optional[bool] = Field(False, description="Clean up extra spaces in text output")
    prefix: Optional[str] = Field(None, description="Prefix added to prompt")
    handle_long_generation: Optional[str] = Field(None, description="Strategy to handle long generation")
    generate_kwargs: Optional[GenerateKwargs] = Field(None, description="Additional kwargs for generate method")

    # Field for conversational model
    messages: Optional[List[Dict[str, str]]] = Field(None, description="Messages for conversational model")

@app.post("/chat")
def generate_text(request_model: UnifiedRequestModel):
    user_generate_kwargs = request_model.generate_kwargs.dict() if request_model.generate_kwargs else {}
    generate_kwargs = {**default_generate_config, **user_generate_kwargs}

    if args.pipeline == "text-generation":
        if not request_model.prompt:
            raise HTTPException(status_code=400, detail="Text generation parameter prompt required")
        sequences = pipeline(
            request_model.prompt,
            # return_tensors=request_model.return_tensors,
            # return_text=request_model.return_text,
            return_full_text=request_model.return_full_text,
            clean_up_tokenization_spaces=request_model.clean_up_tokenization_spaces,
            prefix=request_model.prefix,
            handle_long_generation=request_model.handle_long_generation,
            **generate_kwargs
        )

        result = ""
        for seq in sequences:
            print(f"Result: {seq['generated_text']}")
            result += seq['generated_text']

        return {"Result": result}

    elif args.pipeline == "conversational": 
        if not request_model.messages:
            raise HTTPException(status_code=400, detail="Conversational parameter messages required")

        response = pipeline(
            request_model.messages, 
            clean_up_tokenization_spaces=request_model.clean_up_tokenization_spaces,
            **generate_kwargs
        )
        return {"Result": str(response[-1])}

    else:
        raise HTTPException(status_code=400, detail="Invalid pipeline type")

@app.get("/metrics")
def get_metrics():
    try:
        gpus = GPUtil.getGPUs()
        gpu_info = []
        for gpu in gpus:
            gpu_info.append({
                "id": gpu.id,
                "name": gpu.name,
                "load": f"{gpu.load * 100:.2f}%",  # Format as percentage
                "temperature": f"{gpu.temperature} C",
                "memory": {
                    "used": f"{gpu.memoryUsed / 1024:.2f} GB",
                    "total": f"{gpu.memoryTotal / 1024:.2f} GB"
                }
            })
        return {"gpu_info": gpu_info}
    except Exception as e:
        return {"error": str(e)}

@app.get("/version")
def health_check():
    with open("/workspace/tfs/model_name.txt", "r") as f:
        model_name = f.read()

    return {"version": model_name}

if __name__ == "__main__":
    local_rank = int(os.environ.get("LOCAL_RANK", 0)) # Default to 0 if not set
    port = 5000 + local_rank # Adjust port based on local rank
    uvicorn.run(app=app, host='0.0.0.0', port=port)
