# E2E Fine-Tuning Dataset Files

## Overview

This dataset file is used for conducting end-to-end (E2E) testing for fine-tuning. The Dockerfile builds an image incorporating the [dolly-15k-oai-style](https://huggingface.co/datasets/philschmid/dolly-15k-oai-style) dataset which is then used within an init container specifically for fine-tuning. 

## Files

- **Dockerfile**: Builds the Docker image for the E2E tests.

- **dataset.parquet**: The dataset itself, downloaded from [dolly-15k-oai-style](https://huggingface.co/datasets/philschmid/dolly-15k-oai-style)


## Usage

Build the Docker image with the following command:

```bash

make docker-build-dataset
 