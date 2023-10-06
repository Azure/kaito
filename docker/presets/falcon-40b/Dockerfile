# Use the NVIDIA PyTorch image as a base
FROM nvcr.io/nvidia/pytorch:23.06-py3

# Set the working directory
WORKDIR /workspace/falcon

# First, copy just the requirements.txt file and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY pkg/presets/falcon/requirements.txt /workspace/falcon/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Download the model and tokenizer
RUN python3 -c "from transformers import AutoModelForCausalLM, AutoTokenizer; \
    model_name = 'tiiuae/falcon-40b-instruct'; \
    AutoModelForCausalLM.from_pretrained(model_name, trust_remote_code=True); \
    AutoTokenizer.from_pretrained(model_name)"

# Copy the entire 'presets/falcon' folder to the working directory
COPY pkg/presets/falcon /workspace/falcon
