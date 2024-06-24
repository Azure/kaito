# Kaito Tuning Workspace API

This guide provides instructions on how to use the Kaito Tuning Workspace 
API for parameter-efficient fine-tuning (PEFT) of models. The API supports 
methods like LoRA and QLoRA and allows users to specify their own datasets and 
configuration settings.

## Getting Started

To use the Kaito Tuning Workspace API, you need to define a Workspace custom resource (CR). 
Below are examples of how to define the CR and its various components.

## Example Workspace Definitions
Here are three examples of using the API to define a workspace for tuning different models:

Example 1: Tuning [`phi-3-mini`](../../examples/fine-tuning/kaito_workspace_tuning_phi_3.yaml)

Example 2: Tuning `falcon-7b`
```yaml
apiVersion: kaito.sh/v1alpha1
kind: Workspace
metadata:
  name: workspace-tuning-falcon
resource:
  instanceType: "Standard_NC6s_v3"
  labelSelector:
    matchLabels:
      app: tuning-phi-3-falcon
tuning:
  preset:
    name: falcon-7b
  method: qlora
  input:
    image: ACR_REPO_HERE.azurecr.io/IMAGE_NAME_HERE:0.0.1
    imagePullSecrets: 
      - IMAGE_PULL_SECRETS_HERE
  output:
    image: ACR_REPO_HERE.azurecr.io/IMAGE_NAME_HERE:0.0.1  # Tuning Output
    imagePushSecret: aimodelsregistrysecret

```
Generic TuningSpec Structure: 
```yaml
tuning:
  preset:
    name: preset-model
  method: lora or qlora
  config: custom-configmap (optional)
  input: # Image or URL
    urls:
      - "https://example.com/dataset.parquet?download=true"
  output: # Image
    image: "youracr.azurecr.io/custom-adapter:0.0.1"
    imagePushSecret: youracrsecret
```

## Default ConfigMaps
The default configuration for different tuning methods can be specified 
using ConfigMaps. Below are examples for LoRA and QLoRA methods.

[Default LoRA ConfigMap](../../charts/kaito/workspace/templates/lora-params.yaml)

[QLoRA ConfigMap](../../charts/kaito/workspace/templates/qlora-params.yaml)

## Using Custom ConfigMaps
You can specify your own custom ConfigMap and include it in the `Config` 
field of the `TuningSpec`

For more information on configurable parameters, please refer to the respective 
documentation links provided in the default ConfigMap examples.
