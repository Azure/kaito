
## Step 1: Clone the Repository

```
git clone https://github.com/Azure/kaito.git
```

## Step 2: Download Your Private/Custom Model Weights

For example, assuming HuggingFace weights:
```
git lfs install
git clone git@hf.co:<MODEL_ID>  # Example: git clone git@hf.co:bigscience/bloom
# OR
git clone https://huggingface.co/bigscience/bloom
```

Alternatively, use curl:
```
curl -sSL https://huggingface.co/bigscience/bloom/resolve/main/config.json?download=true -o config.json
```
More Info [Here](https://huggingface.co/docs/hub/en/models-downloading)


## Step 3: Build Docker Image with Private/Custom Weights

Navigate to the Kaito base directory and build the Docker image, including the weights directory in the build context:

```
docker build -t <IMAGE_NAME> --file docker/presets/models/tfs/Dockerfile --build-arg WEIGHTS_PATH=<WEIGHTS_PATH> --build-arg MODEL_TYPE=text-generation --build-arg VERSION=0.0.1 .

docker push <IMAGE_NAME>
```

- Example Image Name: `modelsregistry.azurecr.io/phi-3-mini-4k-instruct:0.0.1`
- Example Weights Path: `kaito/phi-3-mini-4k-instruct/weights`


## Step 4: Deploy
Follow the [Custom Template](./custom-hf-model-guide.md)
