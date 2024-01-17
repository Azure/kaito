# Use the MCR PyTorch image as a base
FROM mcr.microsoft.com/aifx/acpt/stable-ubuntu2004-cu118-py38-torch211

ARG WEIGHTS_PATH
ARG MODEL_PRESET_PATH

# Set the working directory
WORKDIR /workspace/tfs

# First, copy just the requirements.txt file and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY presets/models/requirements.txt /workspace/tfs/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY presets/models/text-gen-inference.py /workspace/tfs/text-gen-inference.py

# Convert to ONNX Runtime
# RUN python convert_to_onnx.py ${MODEL_NAME} 

# Copy the entire model weights to the weights directory
COPY ${WEIGHTS_PATH} /workspace/tfs/weights
