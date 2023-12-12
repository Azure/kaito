# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
# System
import os
import argparse

# API
from typing import List, Optional
from pydantic import BaseModel, Field
from fastapi import FastAPI, HTTPException
import uvicorn

# ML
from transformers import AutoTokenizer, AutoModelForCausalLM
import transformers
import torch

def dtype_type(string):
    if hasattr(torch, string):
        return getattr(torch, string)
    else:
        raise ValueError(f"Invalid torch dtype: {string}")

parser = argparse.ArgumentParser(description='Model Configuration')
parser.add_argument('--model_id', required=True, type=str, help='The model ID for the pre-trained model')
parser.add_argument('--pipeline', required=True, type=str, help='The model pipeline for the pre-trained model')
parser.add_argument('--load_in_8bit', default=False, action='store_true', help='Load model in 8-bit mode')
parser.add_argument('--trust_remote_code', default=False, action='store_true', help='Disable trusting remote code when loading the model')
parser.add_argument('--torch_dtype', default=None, type=dtype_type, help='The torch dtype for the pre-trained model')
parser.add_argument('--device_map', default="auto", type=str, help='The device map for the pre-trained model')

args = parser.parse_args()

app = FastAPI()

supported_pipelines = {"conversational", "text-generation"}
if args.pipeline not in supported_pipelines:
    raise HTTPException(status_code=400, detail="Invalid pipeline specified")

model_kwargs = {
    "device_map": args.device_map,
    "trust_remote_code": args.trust_remote_code,
}

if args.load_in_8bit:
    model_kwargs["load_in_8bit"] = args.load_in_8bit
if args.torch_dtype:
    model_kwargs["torch_dtype"] = args.torch_dtype

tokenizer = AutoTokenizer.from_pretrained("/workspace/tfs/weights")
model = AutoModelForCausalLM.from_pretrained(
    "/workspace/tfs/weights",
    **model_kwargs
)

pipeline_kwargs = {
    "trust_remote_code": args.trust_remote_code,
}

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
    max_length: Optional[int] = Field(200, description="Maximum length for generated text")
    min_length: Optional[int] = Field(0, description="Minimum length for generated text")
    do_sample: Optional[bool] = Field(True, description="Whether to use sampling; set to False to use greedy decoding")
    early_stopping: Optional[bool] = Field(False, description="Whether to stop the model when it produces the EOS token")
    num_beams: Optional[int] = Field(1, description="Number of beams for beam search")
    num_beam_groups: Optional[int] = Field(1, description="Number of groups for diverse beam search")
    diversity_penalty: Optional[float] = Field(0.0, description="Diversity penalty for diverse beam search")
    temperature: Optional[float] = Field(1.0, description="Temperature for sampling distribution")
    top_k: Optional[int] = Field(10, description="The number of highest probability vocabulary tokens to keep for top-k-filtering")
    top_p: Optional[float] = Field(1.0, description="Nucleus filtering (top-p) threshold")
    typical_p: Optional[float] = Field(1.0, description="Typical (set to 1 to ignore typical_p sampling)")
    repetition_penalty: Optional[float] = Field(1.0, description="Parameter for repetition penalty")
    length_penalty: Optional[float] = Field(1.0, description="Exponential penalty to the length")
    no_repeat_ngram_size: Optional[int] = Field(0, description="Size of the no repeat n-gram")
    encoder_no_repeat_ngram_size: Optional[int] = Field(0, description="Size of the no repeat n-gram in the encoder")
    bad_words_ids: Optional[List[int]] = Field(None, description="List of token ids that are not allowed to be generated")
    num_return_sequences: Optional[int] = Field(1, description="Number of sequences to return")
    output_scores: Optional[bool] = Field(False, description="Whether to return the model's output scores")
    return_dict_in_generate: Optional[bool] = Field(False, description="Whether to return a dictionary instead of a list")
    pad_token_id: Optional[int] = Field(tokenizer.pad_token_id, description="Pad token id")
    eos_token_id: Optional[int] = Field(tokenizer.eos_token_id, description="End of sentence token id")
    forced_bos_token_id: Optional[int] = Field(None, description="Forced beginning of sentence token id")
    forced_eos_token_id: Optional[int] = Field(None, description="Forced end of sentence token id")
    remove_invalid_values: Optional[bool] = Field(None, description="Whether to remove invalid values")

    # Field for conversational model
    messages: Optional[List[dict]] = Field(None, description="Messages for conversational model")

@app.post("/chat")
def generate_text(request_model: UnifiedRequestModel):
    if args.pipeline == "text-generation":
        if not request_model.prompt:
            raise HTTPException(status_code=400, detail="Text generation parameter prompt required")
        sequences = pipeline(
            request_model.prompt,
            max_length=request_model.max_length,
            min_length=request_model.min_length,
            do_sample=request_model.do_sample,
            early_stopping=request_model.early_stopping,
            num_beams=request_model.num_beams,
            num_beam_groups=request_model.num_beam_groups,
            diversity_penalty=request_model.diversity_penalty,
            temperature=request_model.temperature,
            top_k=request_model.top_k,
            top_p=request_model.top_p,
            typical_p=request_model.typical_p,
            repetition_penalty=request_model.repetition_penalty,
            length_penalty=request_model.length_penalty,
            no_repeat_ngram_size=request_model.no_repeat_ngram_size,
            encoder_no_repeat_ngram_size=request_model.encoder_no_repeat_ngram_size,
            bad_words_ids=request_model.bad_words_ids,
            num_return_sequences=request_model.num_return_sequences,
            output_scores=request_model.output_scores,
            return_dict_in_generate=request_model.return_dict_in_generate,
            pad_token_id=request_model.pad_token_id,
            eos_token_id=request_model.eos_token_id,
            forced_bos_token_id=request_model.forced_bos_token_id,
            forced_eos_token_id=request_model.forced_eos_token_id,
            remove_invalid_values=request_model.remove_invalid_values,
        )

        result = ""
        for seq in sequences:
            print(f"Result: {seq['generated_text']}")
            result += seq['generated_text']

        return {"Result": result}
    
    elif args.pipeline == "conversational": 
        if not request_model.messages:
            raise HTTPException(status_code=400, detail="Conversational parameter messages required")
        
        response = pipeline(request_model.messages)
        return {"Result": str(response[-1])}
    
    else:
        raise HTTPException(status_code=400, detail="Invalid pipeline type")


if __name__ == "__main__":
    local_rank = int(os.environ.get("LOCAL_RANK", 0)) # Default to 0 if not set
    port = 5000 + local_rank # Adjust port based on local rank
    uvicorn.run(app=app, host='0.0.0.0', port=port)
