FROM busybox:latest

RUN mkdir -p /mnt/adapters

COPY docker/adapter/adapter_config.json /mnt/adapters/
COPY docker/adapter/adapter_model.safetensors /mnt/adapters/