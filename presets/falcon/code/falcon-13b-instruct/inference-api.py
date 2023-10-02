# !pip install -U bitsandbytes
# !pip install -U git+https://github.com/huggingface/transformers.git
# !pip install -U git+https://github.com/huggingface/accelerate.git
# !pip install fastapi pydantic
# !pip install 'uvicorn[standard]'

# System
import os

# API
from fastapi import FastAPI, HTTPException
import uvicorn

# ML
from transformers import AutoTokenizer, AutoModelForCausalLM
import transformers
import torch
import torch.distributed as dist

app = FastAPI()
model_id = "tiiuae/falcon-13b-instruct"

tokenizer = AutoTokenizer.from_pretrained(model_id)
model = AutoModelForCausalLM.from_pretrained(
    model_id,
    device_map="auto",
    torch_dtype=torch.bfloat16,
    trust_remote_code=True,
    # offload_folder="offload",
    # offload_state_dict = True
    # load_in_8bit=True,
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
    
@app.post("/generate")
def generate_text(prompt: str):
    sequences = pipeline(
        prompt,
        max_length=200,
        do_sample=True,
        top_k=10,
        num_return_sequences=1,
        eos_token_id=tokenizer.eos_token_id,
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
