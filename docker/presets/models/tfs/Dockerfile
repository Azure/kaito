FROM python:3.10-slim@sha256:684b1aaf96a7942b3c3af438d162e0baa3510aa7af25ad76d238e0c746bdec79

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
# Inference
COPY kaito/presets/inference/${MODEL_TYPE}/requirements.txt /workspace/tfs/inference-requirements.txt
RUN pip install --no-cache-dir -r inference-requirements.txt

COPY kaito/presets/inference/${MODEL_TYPE}/inference_api.py /workspace/tfs/inference_api.py

# Fine Tuning
COPY kaito/presets/tuning/${MODEL_TYPE}/requirements.txt /workspace/tfs/tuning-requirements.txt
RUN pip install --no-cache-dir -r tuning-requirements.txt

COPY kaito/presets/tuning/${MODEL_TYPE}/cli.py /workspace/tfs/cli.py
COPY kaito/presets/tuning/${MODEL_TYPE}/fine_tuning_api.py /workspace/tfs/fine_tuning_api.py
COPY kaito/presets/tuning/${MODEL_TYPE}/parser.py /workspace/tfs/parser.py
COPY kaito/presets/tuning/${MODEL_TYPE}/dataset.py /workspace/tfs/dataset.py

# Copy the entire model weights to the weights directory
COPY ${WEIGHTS_PATH} /workspace/tfs/weights
