# System
import os
import argparse

# API
from typing import List, Optional
from pydantic import BaseModel
from fastapi import FastAPI, HTTPException
import uvicorn

# ML
from transformers import AutoTokenizer, AutoModelForCausalLM
import transformers
import torch
import torch.distributed as dist

parser = argparse.ArgumentParser(description='Falcon Model Configuration')
parser.add_argument('--load_in_8bit', default=False, action='store_true', help='Load model in 8-bit mode')
parser.add_argument('--disable_trust_remote_code', default=False, action='store_true', help='Disable trusting remote code when loading the model')
# parser.add_argument('--model_id', required=True, type=str, help='The Falcon ID for the pre-trained model')
args = parser.parse_args()

app = FastAPI()

tokenizer = AutoTokenizer.from_pretrained("/workspace/falcon/weights")
model = AutoModelForCausalLM.from_pretrained(
    "/workspace/falcon/weights", # args.model_id,
    device_map="auto",
    torch_dtype=torch.bfloat16,
    trust_remote_code=not args.disable_trust_remote_code, # Use NOT since our flag disables the trust
    load_in_8bit=args.load_in_8bit,
    # offload_folder="offload",
    # offload_state_dict = True
)

pipeline = transformers.pipeline(
    "text-generation",
    model=model,
    tokenizer=tokenizer,
    torch_dtype=torch.bfloat16,
    trust_remote_code=True,
    device_map="auto",
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

class GenerationParams(BaseModel):
    prompt: str
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
    bad_words_ids: List[int] = None
    num_return_sequences: int = 1
    output_scores: bool = False
    return_dict_in_generate: bool = False
    pad_token_id: Optional[int] = tokenizer.pad_token_id
    eos_token_id: Optional[int] = tokenizer.eos_token_id
    forced_bos_token_id: Optional[int] = None
    forced_eos_token_id: Optional[int] = None
    remove_invalid_values: Optional[bool] = None


@app.post("/chat")
def generate_text(params: GenerationParams):
    sequences = pipeline(
        params.prompt,
        max_length=params.max_length,
        min_length=params.min_length,
        do_sample=params.do_sample,
        early_stopping=params.early_stopping,
        num_beams=params.num_beams,
        num_beam_groups=params.num_beam_groups,
        diversity_penalty=params.diversity_penalty,
        temperature=params.temperature,
        top_k=params.top_k,
        top_p=params.top_p,
        typical_p=params.typical_p,
        repetition_penalty=params.repetition_penalty,
        length_penalty=params.length_penalty,
        no_repeat_ngram_size=params.no_repeat_ngram_size,
        encoder_no_repeat_ngram_size=params.encoder_no_repeat_ngram_size,
        bad_words_ids=params.bad_words_ids,
        num_return_sequences=params.num_return_sequences,
        output_scores=params.output_scores,
        return_dict_in_generate=params.return_dict_in_generate,
        forced_bos_token_id=params.forced_bos_token_id,
        forced_eos_token_id=params.forced_eos_token_id,
        eos_token_id=params.eos_token_id,
        remove_invalid_values=params.remove_invalid_values
    )

    result = ""
    for seq in sequences:
        print(f"Result: {seq['generated_text']}")
        result += seq['generated_text']

    return {"Result": result}


if __name__ == "__main__":
    local_rank = int(os.environ.get("LOCAL_RANK", 0)) # Default to 0 if not set
    port = 5000 + local_rank # Adjust port based on local rank
    uvicorn.run(app=app, host='0.0.0.0', port=port)
