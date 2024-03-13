FROM python:3.10-slim

ARG WEIGHTS_PATH
ARG MODEL_TYPE
ARG VERSION
ARG IMAGE_NAME

# Set the working directory
WORKDIR /workspace/tfs

# Write metadata to model_info.txt file
RUN echo -e "$VERSION\n$IMAGE_NAME" > /workspace/tfs/model_info.txt

# First, copy just the preset files and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY kaito/presets/inference/${MODEL_TYPE}/requirements.txt /workspace/tfs/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY kaito/presets/inference/${MODEL_TYPE}/inference_api.py /workspace/tfs/inference_api.py

# Copy the entire model weights to the weights directory
COPY ${WEIGHTS_PATH} /workspace/tfs/weights
