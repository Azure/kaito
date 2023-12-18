FROM nvcr.io/nvidia/pytorch:23.10-py3

ARG WEIGHTS_PATH
ARG MODEL_PRESET_PATH

# Set the working directory
WORKDIR /workspace/tfs

# First, copy just the requirements.txt file and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY ${MODEL_PRESET_PATH}/requirements.txt /workspace/tfs/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Copy the entire model weights to the weights directory
COPY ${WEIGHTS_PATH} /workspace/tfs/weights

# Copy the entire 'presets' folder to the working directory
COPY ${MODEL_PRESET_PATH} /workspace/tfs