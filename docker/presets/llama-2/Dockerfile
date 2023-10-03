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

RUN git clone https://github.com/facebookresearch/llama

WORKDIR /workspace/llama

RUN pip install -e .
RUN pip install fastapi pydantic
RUN pip install 'uvicorn[standard]'

ARG LLAMA_VERSION
ARG SRC_DIR

# Conditional logic based on LLAMA_VERSION argument
RUN if [ "$LLAMA_VERSION" = "llama-2-7b" ]; then \
      wget -P ${SRC_DIR}/weights <URL_for_7b_chat_model>; \
    elif [ "$LLAMA_VERSION" = "llama-2-13b" ]; then \
      wget -P ${SRC_DIR}/weights <URL_for_13b_chat_model>; \
    elif [ "$LLAMA_VERSION" = "llama-2-70b" ]; then \
      wget -P ${SRC_DIR}/weights <URL_for_70b_chat_model>; \
    elif [ "$LLAMA_VERSION" = "llama-2-7b-chat" ]; then \
      wget -P ${SRC_DIR}/weights <URL_for_7b_model>; \
    elif [ "$LLAMA_VERSION" = "llama-2-13b-chat" ]; then \
      wget -P ${SRC_DIR}/weights <URL_for_13b_model>; \
    elif [ "$LLAMA_VERSION" = "llama-2-70b-chat" ]; then \
      wget -P ${SRC_DIR}/weights <URL_for_70b_model>; \
    else \
      echo "Invalid or missing LLAMA_VERSION"; \
      exit 1; \
    fi

ADD ${SRC_DIR} /workspace/llama/llama-2
