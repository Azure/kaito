# Build Llama model
# docker build -t $ACR_NAME.azurecr.io/{{IMAGE_NAME}}:$IMAGE_TAG \
#              --build-arg WEIGHTS_PATH=/weights \
#              --build-arg VERSION={{VERSION}} \
#              --build-arg MODEL_TYPE={{MODEL_TYPE}} \
#              --build-arg IMAGE_NAME={{IMAGE_NAME}} \
#              --build-arg MODEL_VERSION={{MODEL_VERSION}} \

FROM python:3.8-slim
WORKDIR /workspace

# Install git
RUN apt-get update && \
    apt-get install -y git && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
    
RUN git clone https://github.com/facebookresearch/llama.git

WORKDIR /workspace/llama

RUN ["/bin/bash", "-c", "sed -i $'/torch.distributed.init_process_group(\"nccl\")/c\\            import datetime\\\n            torch.distributed.init_process_group(\"nccl\", timeout=datetime.timedelta(days=365*100))' /workspace/llama/llama/generation.py"]

RUN pip install -e .
RUN pip install torch==2.1.0 fastapi==0.109.1 pydantic==1.10.9 gputil==1.4.0
RUN pip install 'uvicorn[standard]'

ARG WEIGHTS_PATH
ARG MODEL_TYPE
ARG VERSION
ARG IMAGE_NAME
ARG MODEL_VERSION

# Write metadata to model_info.json file
RUN MODEL_VERSION_HASH="${MODEL_VERSION##*/}" && \
    echo "{\"Model Type\": \"$MODEL_TYPE\", \"Version\": \"$VERSION\", \"Image Name\": \"$IMAGE_NAME\", \"Model Version URL\": \"$MODEL_VERSION\", \"REVISION_ID\": \"$MODEL_VERSION_HASH\"}" > /workspace/llama/model_info.json

ADD ${WEIGHTS_PATH} /workspace/llama/llama-2/weights
ADD kaito/presets/inference/${MODEL_TYPE} /workspace/llama/llama-2
