# Build text completion model
# docker build \
#   --build-arg LLAMA_WEIGHTS=llama-2-7b \
#   --build-arg SRC_DIR=/home/presets/llama-2 \
#   -t llama-2-7b:latest .

# Build chat completion model
# docker build \
#   --build-arg LLAMA_WEIGHTS=llama-2-7b-chat \
#   --build-arg SRC_DIR=/home/presets/llama-2-chat \
#   -t llama-2-7b-chat:latest .

FROM nvcr.io/nvidia/pytorch:23.10-py3
WORKDIR /workspace

RUN git clone https://github.com/facebookresearch/llama

WORKDIR /workspace/llama

RUN sed -i $'/torch.distributed.init_process_group("nccl")/c\\            import datetime\\\n            torch.distributed.init_process_group("nccl", timeout=datetime.timedelta(days=365*100))' /workspace/llama/llama/generation.py

RUN pip install -e .
RUN pip install fastapi pydantic
RUN pip install 'uvicorn[standard]'

ARG WEIGHTS_PATH
ARG MODEL_PRESET_PATH

ADD ${WEIGHTS_PATH} /workspace/llama/llama-2/weights
ADD ${MODEL_PRESET_PATH} /workspace/llama/llama-2
