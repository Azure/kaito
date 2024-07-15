# Kaito Inference Workspace API

This guide provides instructions on how to use the Kaito Inference Workspace API for basic model serving and serving with LoRA adapters.

## Getting Started

To use the Kaito Inference Workspace API, you need to define a Workspace custom resource (CR). Below are examples of how to define the CR and its various components.

## Example Workspace Definitions
Here are three examples of using the API to define a workspace for inferencing different models:

Example 1: Inferencing [`phi-3-mini`](../../examples/inference/kaito_workspace_phi_3.yaml)

Example 2: Inferencing [`falcon-7b`](../../examples/inference/kaito_workspace_falcon_7b.yaml) without adapters

Example 3: Inferencing `falcon-7b` with adapters

```yaml
apiVersion: kaito.sh/v1alpha1
kind: Workspace
metadata:
  name: workspace-falcon-7b
resource:
  instanceType: "Standard_NC12s_v3"
  labelSelector:
    matchLabels:
      apps: falcon-7b
inference:
  preset:
    name: "falcon-7b"
  adapters:
    - source:
        name: "falcon-7b-adapter"
        image:  "<YOUR_IMAGE>"
      strength: "0.2"
```

Multiple adapters can be added:

```yaml
apiVersion: kaito.sh/v1alpha1
kind: Workspace
metadata:
  name: workspace-falcon-7b
resource:
  instanceType: "Standard_NC12s_v3"
  labelSelector:
    matchLabels:
      apps: falcon-7b
inference:
  preset:
    name: "falcon-7b"
  adapters:
    - source:
        name: "falcon-7b-adapter"
        image:  "<YOUR_IMAGE>"
      strength: "0.2"
    - source:
        name: "additional-source"
        image: "<YOUR_ADDITIONAL_IMAGE>"
      strength: "0.5" 
```

Currently, only images are supported as adapter sources, with a default strength of "1.0".