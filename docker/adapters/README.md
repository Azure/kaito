# E2E Adapter Test Files

## Overview

These files are part of a set used for conducting end-to-end (E2E) testing of an adapter component. The Dockerfile builds an image incorporating the configuration and model files, which is then used within an Init Container for testing. The adapter is training from [dolly-15k-oai-style](https://huggingface.co/datasets/philschmid/dolly-15k-oai-style) dataset
and was trained using default [qlora-params.yaml](../../charts/kaito/workspace/templates/qlora-params.yaml) 

## Files

- **Dockerfile**: Builds the Docker image for the E2E tests.

- **adapter_config.json**: Contains settings for configuring the adapter in the test environment.

- **adapter_model.safetensors**: Provides the adapter's machine learning model in SafeTensors format.

## Usage

Build the Docker image with the following command:

```bash

make docker-build-adapter
 