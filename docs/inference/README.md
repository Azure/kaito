# Kaito Inference

This document presents how to use the Kaito `workspace` Custom Resource Definition (CRD) for model serving and serving with LoRA adapters.

## Usage

The basic usage for inference is simple. Users just need to specify the GPU SKU used for inference in the `resource` spec and one of the Kaito supported model name in the `inference` spec in the `workspace` custom resource. For example,

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
```

If a user runs Kaito in an on-premise Kubernetes cluster where GPU SKUs are unavailable, the GPU nodes can be prepared in advance. The user should ensure the corresponding vendor specific GPU plugin is installed successfully in every prepared node, i.e., the node status should report a non-zero allocatable GPU resource count. For example,

```
$ kubectl get node $NODE_NAME -o json | jq .status.allocatable
{
  "cpu": "XXXX",
  "ephemeral-storage": "YYYY",
  "hugepages-1Gi": "0",
  "hugepages-2Mi": "0",
  "memory": "ZZZZ",
  "nvidia.com/gpu": "1",
  "pods": "100"
}
```

Next, the user needs to add the node names in the `preferredNodes` field in the `resource` spec. As a result, the Kaito controller will skip the steps for GPU node provisioning and use the prepared nodes to run the inference workload.
> [!IMPORTANT]
> The node objects of the preferred nodes need to contain the same matching labels as specified in the `resource` spec. Otherwise, the Kaito controller would not recognize them.

### Inference with LoRA adapters 

Kaito also supports running the inference workload with LoRA adapters produced by [model fine-tuning jobs](../tuning/README.md). Users can specify one or more adapters in the `adapters` field of the `inference` spec. For example,

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
Currently, only images are supported as adapter sources. The `strength` field specifies the multiplier for applying the adapter weights to the raw model weights.

Note: if you build a container image for an existing adapter, please copy all adapter files to the **/data** directory inside the container.

The detailed `InferenceSpec` API definitions can be found [here](https://github.com/Azure/kaito/blob/2ccc93daf9d5385649f3f219ff131ee7c9c47f3e/api/v1alpha1/workspace_types.go#L75).


# Inference workload

Depending on whether the specified model supports distributed inference or not, the Kaito controller will choose to use either Kubernetes **apps.deployment** workload (by default) or Kubernetes **apps.statefulset** workload (if the model supports distributed inference) to manage the inference service, which is exposed using a Cluster-IP type of Kubernetes `service`.

When adapters are specified in the `inference` spec, the Kaito controller adds an initcontainer for each adapter in addition to the main container. The pod structure is shown in Figure 1.

<div align="left">
  <img src="../img/kaito-inference-adapter.png" width=40% title="Kaito inference adapter" alt="Kaito inference adapter">
</div>

If an image is specified as the adapter source, the corresponding initcontainer uses that image as its container image. These initcontainers ensure all adapter data is available locally before the inference service starts. The main container uses a supported model image, launching the [inference_api.py](https://github.com/Azure/kaito/presets/inference/text-generation/inference_api.py) script.

All containers share local volumes by mounting the same `EmptyDir` volumes, avoiding file copies between containers.

## Workload update

Users can update the `workspace` customer resource and change the `adapters` field in the inference spec. The Kaito controller will apply a workload deployment update to make adapter changes take effect. The inference service pod will be recreated, causing a short period of service downtime. Once the new adapters are merged with the raw model weights and loaded into the GPU memory, the service will resume.


# Troubleshooting

TBD
