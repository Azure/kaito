FROM python:3.10-slim@sha256:684b1aaf96a7942b3c3af438d162e0baa3510aa7af25ad76d238e0c746bdec79

ARG WEIGHTS_PATH
ARG MODEL_TYPE
ARG VERSION

# Set the working directory
WORKDIR /workspace

# Write the version to a file
RUN pip install --no-cache-dir virtualenv

# 1. Huggingface transformers
# dependencies
COPY kaito/presets/inference/${MODEL_TYPE}/requirements.txt /workspace/tfs/inference-requirements.txt
COPY kaito/presets/tuning/${MODEL_TYPE}/requirements.txt /workspace/tfs/tuning-requirements.txt
RUN virtualenv tfs && \
    . tfs/bin/activate && \
    pip install --no-cache-dir -r /workspace/tfs/inference-requirements.txt && \
    pip install --no-cache-dir -r /workspace/tfs/tuning-requirements.txt && \
    deactivate

# Copy the inference and tuning scripts
COPY kaito/presets/inference/${MODEL_TYPE}/inference_api.py /workspace/tfs/inference_api.py
COPY kaito/presets/tuning/${MODEL_TYPE}/cli.py /workspace/tfs/cli.py
COPY kaito/presets/tuning/${MODEL_TYPE}/fine_tuning.py /workspace/tfs/fine_tuning.py
COPY kaito/presets/tuning/${MODEL_TYPE}/parser.py /workspace/tfs/parser.py
COPY kaito/presets/tuning/${MODEL_TYPE}/dataset.py /workspace/tfs/dataset.py

# Copy the metrics server
COPY kaito/presets/tuning/${MODEL_TYPE}/metrics/metrics_server.py /workspace/tfs/metrics_server.py

# 2. vLLM
COPY kaito/presets/inference/vllm/requirements.txt /workspace/vllm/inference-requirements.txt
RUN virtualenv vllm && \
    . vllm/bin/activate && \
    pip install --no-cache-dir -r /workspace/vllm/inference-requirements.txt && \
    deactivate

COPY kaito/presets/inference/vllm/inference_api.py /workspace/vllm/inference_api.py

# 3. Model weights
COPY kaito/docker/presets/models/tfs/entrypoint /workspace/entrypoint
COPY ${WEIGHTS_PATH} /workspace/weights
RUN echo $VERSION > /workspace/version.txt && \
    ln -s /workspace/weights /workspace/tfs/weights && \
    ln -s /workspace/weights /workspace/vllm/weights

ENTRYPOINT ["/workspace/entrypoint"]
