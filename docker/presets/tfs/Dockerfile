FROM python:3.8-slim

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
COPY kaito/presets/inference/${MODEL_TYPE}/requirements.txt /workspace/tfs/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY kaito/presets/inference/${MODEL_TYPE}/inference-api.py /workspace/tfs/inference-api.py

# Copy only files that don't start with a "." from the weights path
RUN find ${WEIGHTS_PATH} -type f ! -name ".*" -exec cp {} /workspace/tfs/weights/ \;
