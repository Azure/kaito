# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import argparse
import os
from typing import Any, Dict, List, Optional

import GPUtil
import torch
import transformers
import uvicorn
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from transformers import AutoModelForCausalLM, AutoTokenizer, GenerationConfig


def dtype_type(string):
    if hasattr(torch, string):
        return getattr(torch, string)
    else:
        raise ValueError(f"Invalid torch dtype: {string}")

parser = argparse.ArgumentParser(description='Model Configuration')
parser.add_argument('--pipeline', required=True, type=str, help='The model pipeline for the pre-trained model')
parser.add_argument('--load_in_8bit', default=False, action='store_true', help='Load model in 8-bit mode')
parser.add_argument('--trust_remote_code', default=False, action='store_true', help='Enable trusting remote code when loading the model')
parser.add_argument('--torch_dtype', default=None, type=dtype_type, help='The torch dtype for the pre-trained model')
parser.add_argument('--device_map', default="auto", type=str, help='The device map for the pre-trained model')
parser.add_argument('--cache_dir', type=str, default=None, help='Cache directory for the model')
parser.add_argument('--from_tf', action='store_true', default=False, help='Load model from a TensorFlow checkpoint')
parser.add_argument('--force_download', action='store_true', default=False, help='Force the download of the model')
parser.add_argument('--resume_download', action='store_true', default=False, help='Resume an interrupted download')
parser.add_argument('--proxies', type=str, default=None, help='Proxy configuration for downloading the model')
parser.add_argument('--revision', type=str, default="main", help='Specific model version to use')
# parser.add_argument('--local_files_only', action='store_true', default=False, help='Only use local files for model loading')
parser.add_argument('--output_loading_info', action='store_true', default=False, help='Output additional loading information')

args = parser.parse_args()

app = FastAPI()

supported_pipelines = {"conversational", "text-generation"}
if args.pipeline not in supported_pipelines:
    raise HTTPException(status_code=400, detail="Invalid pipeline specified")

model_kwargs = {
    "cache_dir": args.cache_dir,
    "from_tf": args.from_tf,
    "force_download": args.force_download,
    "resume_download": args.resume_download,
    "proxies": args.proxies,
    "revision": args.revision,
    "output_loading_info": args.output_loading_info,
    "trust_remote_code": args.trust_remote_code,
    "device_map": args.device_map,
    "local_files_only": True,
}

if args.load_in_8bit:
    model_kwargs["load_in_8bit"] = args.load_in_8bit
if args.torch_dtype:
    model_kwargs["torch_dtype"] = args.torch_dtype

tokenizer = AutoTokenizer.from_pretrained("/workspace/tfs/weights", **model_kwargs)
model = AutoModelForCausalLM.from_pretrained(
    "/workspace/tfs/weights",
    **model_kwargs
)

pipeline_kwargs = {
    "trust_remote_code": args.trust_remote_code,
    "device_map": args.device_map,
}

if args.torch_dtype:
    pipeline_kwargs["torch_dtype"] = args.torch_dtype

pipeline = transformers.pipeline(
    args.pipeline,
    model=model,
    tokenizer=tokenizer,
    **pipeline_kwargs
)

@app.get('/')
def home():
    return "Server is running", 200

@app.get("/healthz")
def health_check():
    if not torch.cuda.is_available():
        raise HTTPException(status_code=500, detail="No GPU available")
    if not model:
        raise HTTPException(status_code=500, detail="Falcon model not initialized")
    if not pipeline: 
        raise HTTPException(status_code=500, detail="Falcon pipeline not initialized")
    return {"status": "Healthy"}

class UnifiedRequestModel(BaseModel):
    # Fields for text generation
    prompt: Optional[str] = Field(None, description="Prompt for text generation")
    # Mutually Exclusive with return_full_text
    # return_tensors: Optional[bool] = Field(False, description="Return tensors of predictions")
    # return_text: Optional[bool] = Field(True, description="Return decoded texts in the outputs")
    return_full_text: Optional[bool] = Field(True, description="Return full text if True, else only added text")
    clean_up_tokenization_spaces: Optional[bool] = Field(False, description="Clean up extra spaces in text output")
    prefix: Optional[str] = Field(None, description="Prefix added to prompt")
    handle_long_generation: Optional[str] = Field(None, description="Strategy to handle long generation")
    generate_kwargs: Optional[Dict[str, Any]] = Field(None, description="Additional kwargs for generate method")

    # Field for conversational model
    messages: Optional[List[dict]] = Field(None, description="Messages for conversational model")

@app.post("/chat")
def generate_text(request_model: UnifiedRequestModel):
    try:
        # Attempt to load the generation configuration
        default_generate_config = GenerationConfig.from_pretrained("/workspace/tfs/weights", local_files_only=True).to_dict()
    except Exception as e:
        default_generate_config = {}
    user_generate_kwargs = request_model.generate_kwargs or {}
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

if __name__ == "__main__":
    local_rank = int(os.environ.get("LOCAL_RANK", 0)) # Default to 0 if not set
    port = 5000 + local_rank # Adjust port based on local rank
    uvicorn.run(app=app, host='0.0.0.0', port=port)
