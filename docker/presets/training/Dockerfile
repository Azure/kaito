FROM python:3.10-slim

ARG WEIGHTS_PATH
ARG MODEL_TYPE
ARG VERSION

# Set the working directory
WORKDIR /workspace/tfs

# Write the version to a file
RUN echo $VERSION > /workspace/tfs/version.txt

# First, copy just the preset files and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY kaito/presets/training/${MODEL_TYPE}/requirements.txt /workspace/tfs/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY kaito/presets/training/${MODEL_TYPE}/cli.py /workspace/tfs/cli.py
COPY kaito/presets/training/${MODEL_TYPE}/training_api.py /workspace/tfs/training_api.py

# Copy the entire model weights to the weights directory
COPY ${WEIGHTS_PATH} /workspace/tfs/weights
