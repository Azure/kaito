# Use the MCR PyTorch image as a base
FROM mcr.microsoft.com/aifx/acpt/stable-ubuntu2004-cu118-py38-torch211

ARG WEIGHTS_PATH
ARG MODEL_TYPE
ARG VERSION
ARG IMAGE_NAME
ARG MODEL_VERSION

# Set the working directory
WORKDIR /workspace/tfs

# Write metadata to model_info.json file
RUN MODEL_VERSION_HASH=$(echo "$MODEL_VERSION" | awk -F'/' '{print $NF}') && \
    echo "{\"Model Type\": \"$MODEL_TYPE\", \"Version\": \"$VERSION\", \"Image Name\": \"$IMAGE_NAME\", \"Model Version URL\": \"$MODEL_VERSION\", \"REVISION_ID\": \"$MODEL_VERSION_HASH\"}" > /workspace/tfs/model_info.json

# First, copy just the requirements.txt file and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY kaito/presets/inference/${MODEL_TYPE}/requirements.txt /workspace/tfs/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY kaito/presets/inference/${MODEL_TYPE}/inference_api.py /workspace/tfs/inference_api.py

# Convert to ONNX Runtime
# RUN python convert_to_onnx.py ${MODEL_NAME} 

# Copy the entire model weights to the weights directory
COPY ${WEIGHTS_PATH} /workspace/tfs/weights
