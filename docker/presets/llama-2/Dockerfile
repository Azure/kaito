# Build text completion model
# docker build \
#   --build-arg LLAMA_VERSION=llama-2-7b \
#   --build-arg SRC_DIR=pkg/presets/llama-2 \
#   -t llama-2-7b:latest .

# Build chat completion model
# docker build \
#   --build-arg LLAMA_VERSION=llama-2-7b-chat \
#   --build-arg SRC_DIR=pkg/presets/llama-2-chat \
#   -t llama-2-7b-chat:latest .

FROM nvcr.io/nvidia/pytorch:23.06-py3
WORKDIR /workspace

# Install Go
RUN apt-get update && apt-get install -y golang-go

RUN git clone https://github.com/facebookresearch/llama

WORKDIR /workspace/llama

RUN pip install -e .
RUN pip install fastapi pydantic
RUN pip install 'uvicorn[standard]'

ARG LLAMA_VERSION
ARG SRC_DIR

# Copy Go download script into the Docker image
COPY docker/presets/llama-2/download_script.go /workspace/download_script.go

# Use Go download script to fetch model weights
RUN go run /workspace/download_script.go $LLAMA_VERSION

ADD ${SRC_DIR} /workspace/llama/llama-2
