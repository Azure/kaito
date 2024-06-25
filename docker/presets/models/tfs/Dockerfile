FROM python:3.10-slim@sha256:684b1aaf96a7942b3c3af438d162e0baa3510aa7af25ad76d238e0c746bdec79

ARG WEIGHTS_PATH
ARG MODEL_TYPE
ARG VERSION

# Set the working directory
WORKDIR /workspace/tfs

# Write the version to a file
RUN echo $VERSION > /workspace/tfs/version.txt

# Install virtualenv
RUN pip install --no-cache-dir virtualenv

# First, create virtual environments for inference and tuning
RUN virtualenv /workspace/tfs/envs/inference
RUN virtualenv /workspace/tfs/envs/tuning

# Copy the requirements files
COPY kaito/presets/inference/${MODEL_TYPE}/requirements.txt /workspace/tfs/inference-requirements.txt
COPY kaito/presets/tuning/${MODEL_TYPE}/requirements.txt /workspace/tfs/tuning-requirements.txt

# Copy in inference file
COPY kaito/presets/inference/${MODEL_TYPE}/inference_api.py /workspace/tfs/inference_api.py

# Install dependencies in their respective virtual environments
RUN /workspace/tfs/envs/inference/bin/pip install --no-cache-dir -r /workspace/tfs/inference-requirements.txt
RUN /workspace/tfs/envs/tuning/bin/pip install --no-cache-dir -r /workspace/tfs/tuning-requirements.txt

# Copy in tuning files
COPY kaito/presets/tuning/${MODEL_TYPE}/cli.py /workspace/tfs/cli.py
COPY kaito/presets/tuning/${MODEL_TYPE}/fine_tuning.py /workspace/tfs/fine_tuning.py
COPY kaito/presets/tuning/${MODEL_TYPE}/parser.py /workspace/tfs/parser.py
COPY kaito/presets/tuning/${MODEL_TYPE}/dataset.py /workspace/tfs/dataset.py

# Copy the entire model weights to the weights directory
COPY ${WEIGHTS_PATH} /workspace/tfs/weights
