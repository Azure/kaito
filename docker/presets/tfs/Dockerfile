FROM nvcr.io/nvidia/pytorch:23.10-py3

ARG MODEL_NAME

# Set the working directory
WORKDIR /workspace/tfs

# First, copy just the requirements.txt file and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY docker/presets/tfs/requirements.txt /workspace/tfs/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Copy the entire model to the weights directory
COPY /home/tfs/${MODEL_NAME} /workspace/tfs/weights
# Copy the entire 'presets/falcon' folder to the working directory
COPY docker/presets/tfs /workspace/tfs