from transformers import AutoTokenizer, AutoModelForCausalLM
import transformers
import torch

model_id = "tiiuae/falcon-7b"

tokenizer = AutoTokenizer.from_pretrained(model_id)
model = AutoModelForCausalLM.from_pretrained(
    model_id,
    device_map="auto",
    torch_dtype=torch.bfloat16,
    trust_remote_code=True,
    # offload_folder="offload",
    # offload_state_dict = True
    # load_in_8bit=True, Uncomment for 8bit quantization
)
pipeline = transformers.pipeline(
    "text-generation",
    model=model,
    tokenizer=tokenizer,
    torch_dtype=torch.bfloat16,
    trust_remote_code=True,
    device_map="auto",
)
sequences = pipeline(
   "Girafatron is obsessed with giraffes, the most glorious animal on the face of this Earth. Giraftron believes all other animals are irrelevant when compared to the glorious majesty of the giraffe.\nDaniel: Hello, Girafatron!\nGirafatron:",
    max_length=200,
    do_sample=True,
    top_k=10,
    num_return_sequences=1,
    eos_token_id=tokenizer.eos_token_id,
)

'''
Without transformers.pipeline
# Initialize the accelerator
accelerator = Accelerator()
print(accelerator.device)
model, tokenizer = accelerator.prepare(model, tokenizer)

input_text = "Girafatron is obsessed with giraffes, the most glorious animal on the face of this Earth. Giraftron believes all other animals are irrelevant when compared to the glorious majesty of the giraffe.\nDaniel: Hello, Girafatron!\nGirafatron:"

# Tokenize the input_text
input_ids = tokenizer.encode(input_text, return_tensors="pt")

# Move input_ids to GPU
input_ids = input_ids.to(accelerator.device)

# Generate sequences
sequences = model.generate(
    input_ids=input_ids,
    max_length=200,
    do_sample=True,
    top_k=10,
    num_return_sequences=1,
    eos_token_id=tokenizer.eos_token_id,
)
'''

for seq in sequences:
    print(f"Result: {seq['generated_text']}")
