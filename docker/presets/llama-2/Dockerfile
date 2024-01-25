# Build Llama model
# docker build -t $ACR_NAME.azurecr.io/{{IMAGE_NAME}}:$IMAGE_TAG \
#              --build-arg WEIGHTS_PATH=/weights \
#              --build-arg VERSION={{VERSION}} \
#              --build-arg MODEL_TYPE={{MODEL_TYPE}} \

FROM nvcr.io/nvidia/pytorch:23.10-py3
WORKDIR /workspace

RUN git clone https://github.com/facebookresearch/llama

WORKDIR /workspace/llama

RUN sed -i $'/torch.distributed.init_process_group("nccl")/c\\            import datetime\\\n            torch.distributed.init_process_group("nccl", timeout=datetime.timedelta(days=365*100))' /workspace/llama/llama/generation.py

RUN pip install -e .
RUN pip install fastapi pydantic
RUN pip install 'uvicorn[standard]'

ARG WEIGHTS_PATH
ARG MODEL_TYPE
ARG VERSION
# Write the version to a file
RUN echo $VERSION > /workspace/llama/version.txt

ADD ${WEIGHTS_PATH} /workspace/llama/llama-2/weights
ADD kaito/presets/models/${MODEL_TYPE} /workspace/llama/llama-2
