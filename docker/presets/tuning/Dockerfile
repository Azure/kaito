FROM python:3.10-slim

ARG WEIGHTS_PATH
ARG MODEL_TYPE
ARG VERSION

# Set the working directory
WORKDIR /workspace/${MODEL_TYPE}

# Write the version to a file
RUN echo $VERSION > /workspace/${MODEL_TYPE}/version.txt

# First, copy just the preset files and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY kaito/presets/tuning/${MODEL_TYPE}/requirements.txt /workspace/${MODEL_TYPE}/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY kaito/presets/tuning/${MODEL_TYPE}/cli.py /workspace/${MODEL_TYPE}/cli.py
COPY kaito/presets/tuning/${MODEL_TYPE}/parser.py /workspace/${MODEL_TYPE}/parser.py
COPY kaito/presets/tuning/${MODEL_TYPE}/fine_tuning_api.py /workspace/${MODEL_TYPE}/fine_tuning_api.py

# Copy the entire model weights to the weights directory
COPY ${WEIGHTS_PATH} /workspace/${MODEL_TYPE}/weights
