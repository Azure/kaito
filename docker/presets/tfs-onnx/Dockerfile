# Use the MCR PyTorch image as a base
FROM mcr.microsoft.com/aifx/acpt/stable-ubuntu2004-cu118-py38-torch211

ARG MODEL_NAME

# Set the working directory
WORKDIR /workspace/tfs

# First, copy just the requirements.txt file and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY /home/presets/tfs-onnx/requirements.txt /workspace/tfs/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Convert to ONNX Runtime
RUN python convert_to_onnx.py ${MODEL_NAME}

# Copy the entire onnx model to the weights directory
COPY /tfs/${MODEL_NAME} /workspace/tfs/weights
# Copy the entire 'presets/tfs-onnx' folder to the working directory
COPY /home/presets/tfs-onnx /workspace/tfs
