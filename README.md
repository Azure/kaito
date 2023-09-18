# KDM

[![Go Report Card](https://goreportcard.com/badge/github.com/Azure/kdm)](https://goreportcard.com/report/github.com/Azure/kdm)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Azure/kdm)

This project introduce `workspace` crd and its controller. The goal is to simplify the workflow of deploying inference services using OSS AI/ML models, and training workloads (to be added) against a standard AKS cluster.

## Quick Start

### Quick Install

Please refer to Helm chart [README](charts/README.md) for more details.

## Demo

1. Create an Azure Kubernetes Service (AKS) cluster

```bash
az group create --name kdm-rg --location eastus
az aks create --name kdm-aks --resource-group kdm-rg --node-count 1  --generate-ssh-keys
```

2. Install [gpu-provisioner](https://github.com/Azure/gpu-provisioner.git) helm chart

```bash

git clone https://github.com/Azure/gpu-provisioner.git
cd gpu-vmprovisioner

AZURE_SUBSCRIPTION_ID=<your_subscription_id> AZURE_LOCATION=<Azure_region> \
AZURE_RESOURCE_GROUP=<your_resource_group_name> AZURE_ACR_NAME=<you_Azure_container_registry_name> \
AZURE_CLUSTER_NAME=<you_AKS_cluster_name> make az-perm az-patch-skaffold-kubenet az-run
```
3. Build and push docker image

```bash
export REGISTRY=<your_docker_registry>
export IMG_NAME=kdm

make docker-build-kdm
```
4. Install KDM helm chart

```bash
helm install kdm --set image.repository=${REGISTRY}/$(IMG_NAME) ./charts/kdm
```

5. Run KDM workspace example

```bash
kubectl apply -f examples/kdm_workspace_llama2_7b-chat.yaml
```

6. Watch the KDM workspace CR status

```bash
watch kubectl describe workspace workspace-llama-2-7b-chat 
```

<details>
<summary>Workspace status</summary>

```bash
Name:         workspace-llama-2-7b-aks
Annotations:  kubernetes-kdm.io/service-type: load-balancer
API Version:  kdm.io/v1alpha1
Inference:
  Preset:
    Name:  llama-2-7b
    Volume:
      Empty Dir:
        Medium:  Memory
      Name:      dshm
Kind:            Workspace
Metadata:
  Creation Timestamp:  2023-09-01T16:41:16Z
  Generation:          1
  Resource Version:    5715733
  UID:                 95db1c71-6a87-408e-96e8-91dc7ef820fd
Resource:
  Count:          2
  Instance Type:  Standard_NC12s_v3
  Label Selector:
    Match Labels:
      apps:  llama-2-7b
  Preferred Nodes:
    node1
    aks-machine98722-26559722-vmss000001
Status:
  Condition:
    Last Transition Time:  2023-09-01T16:41:16Z
    Message:               machine has been provisioned successfully
    Observed Generation:   1
    Reason:                machineProvisionSuccess
    Status:                True
    Type:                  MachineProvisioned
    Last Transition Time:  2023-09-01T16:45:00Z
    Message:               machines plugins have been installed successfully
    Observed Generation:   1
    Reason:                installNodePluginsSuccess
    Status:                True
    Type:                  MachineReady
    Last Transition Time:  2023-09-01T16:45:00Z
    Message:               node plugins have been installed
    Observed Generation:   1
    Reason:                InstallNodePluginsSuccess
    Status:                True
    Type:                  NodePluginsInstalled
    Last Transition Time:  2023-09-01T16:45:00Z
    Message:               workspace resource is ready
    Observed Generation:   1
    Reason:                workspaceResourceDeployedSuccess
    Status:                True
    Type:                  ResourceProvisioned
    Last Transition Time:  2023-09-01T16:45:00Z
    Message:               workspace is ready
    Observed Generation:   1
    Reason:                workspaceReady
    Status:                True
    Type:                  WorkspaceReady
  Worker Nodes:
    aks-machine98722-26559722-vmss000001
    aks-machine13355-19479027-vmss000000
Events:  <none>
```
</details><br/>

7. Clean up

```bash
az aks delete --name kdm-aks --resource-group kdm-rg
```

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
