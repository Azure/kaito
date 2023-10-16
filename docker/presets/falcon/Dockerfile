# Use the NVIDIA PyTorch image as a base
FROM nvcr.io/nvidia/pytorch:23.06-py3

# Set the working directory
WORKDIR /workspace/falcon

# Install Go
RUN apt-get update && apt-get install -y --no-install-recommends golang-go

# First, copy just the requirements.txt file and install dependencies
# This is done before copying the code to utilize Docker's layer caching and
# avoid reinstalling dependencies unless the requirements file changes.
COPY /home/presets/falcon/requirements.txt /workspace/falcon/requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

ARG FALCON_MODEL_NAME

COPY /home/docker/presets/download_script.go /workspace/download_script.go
RUN go run /workspace/download_script.go "public" ${FALCON_MODEL_NAME} "/workspace/falcon/weights"

# Copy the entire 'presets/falcon' folder to the working directory
COPY /home/presets/falcon /workspace/falcon
