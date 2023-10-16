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
RUN apt-get update && apt-get install -y --no-install-recommends golang-go

RUN git clone https://github.com/facebookresearch/llama

WORKDIR /workspace/llama

RUN pip install -e .
RUN pip install fastapi pydantic
RUN pip install 'uvicorn[standard]'

ARG LLAMA_VERSION
ARG SRC_DIR
# ARG EXTERNAL_IP
# ARG EXTERNAL_PORT
# ARG WEB_SERVER_AUTH_TOKEN

# ENV AUTH_TOKEN_ENV_VAR=${WEB_SERVER_AUTH_TOKEN}

# Copy Go download script into the Docker image
# COPY /home/docker/presets/download_script.go /workspace/download_script.go

# Use Go download script to fetch model weights
# RUN go run /workspace/download_script.go "private" ${LLAMA_VERSION} "/workspace/llama/llama-2/weights" ${EXTERNAL_IP} ${EXTERNAL_PORT}
ADD /llama/${LLAMA_VERSION} /workspace/llama/llama-2/weights
ADD ${SRC_DIR} /workspace/llama/llama-2
