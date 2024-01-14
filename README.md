# Kubernetes AI Toolchain Operator (Kaito)

[![Go Report Card](https://goreportcard.com/badge/github.com/Azure/kaito)](https://goreportcard.com/report/github.com/Azure/kaito)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Azure/kaito)
[![codecov](https://codecov.io/gh/Azure/kaito/graph/badge.svg?token=XAQLLPB2AR)](https://codecov.io/gh/Azure/kaito)

| ![notification](docs/img/bell.svg) What is NEW!|
|---------------------------------------------------------------------------------------------|
| First Release: Nov 15th, 2023. Kaito v0.1.0.                                               |

Kaito is an operator that automates the AI/ML inference model deployment in a Kubernetes cluster.
The target models are popular large open sourced inference models such as [falcon](https://huggingface.co/tiiuae) and [llama 2](https://github.com/facebookresearch/llama).
Kaito has the following key differentiations compared to most of the mainstream model deployment methodologies built on top of virtual machine infrastructures.
- Manage large model files using container images. A http server is provided to perform inference calls using the model library.
- Avoid tuning deployment parameters to fit GPU hardware by providing preset configurations.
- Auto-provision GPU nodes based on model requirements.
- Host large model images in public Microsoft Container Registry(MCR) if the license allows.

Using Kaito, the workflow of onboarding large AI inference models in Kubernetes is largely simplified.


## Architecture

Kaito follows the classic Kubernetes Custom Resource Definition(CRD)/controller design pattern. User manages a `workspace` custom resource which describes the GPU requirements and the inference specification. Kaito controllers will automate the deployment by reconciling the `workspace` custom resource.
<div align="left">
  <img src="docs/img/arch.png" width=80% title="Kaito architecture">
</div>

The above figure presents the Kaito architecture overview. Its major components consist of:
- **Workspace controller**: It reconciles the `workspace` custom resource, creates `machine` (explained below) custom resources to trigger node auto provisioning, and creates the inference workload (`deployment` or `statefulset`) based on the model preset configurations.
- **Node provisioner controller**: The controller's name is *gpu-provisioner* in [Kaito helm chart](charts/kaito/gpu-provisioner). It uses the `machine` CRD originated from [Karpenter](https://github.com/aws/karpenter-core) to interact with the workspace controller. It integrates with Azure Kubernetes Service(AKS) APIs to add new GPU nodes to the AKS cluster. 
Note that the *gpu-provisioner* is not an open sourced component. It can be replaced by other controllers if they support Karpenter-core APIs.


## Installation 

The following guidance assumes **Azure Kubernetes Service(AKS)** is used to host the Kubernetes cluster.

Before you begin, ensure you have the following tools installed:

- [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) to provision Azure resources
- [Helm](https://helm.sh) to install this operator
- [kubectl](https://kubernetes.io/docs/tasks/tools/) to view Kubernetes resources
- [git](https://git-scm.com/downloads) to clone this repo locally

If you do not already have an AKS cluster, run the following Azure CLI commands to create one:

```bash
export RESOURCE_GROUP="myResourceGroup"
export MY_CLUSTER="myCluster"
export LOCATION="eastus"
az group create --name $RESOURCE_GROUP --location $LOCATION
az aks create --resource-group $RESOURCE_GROUP --name $MY_CLUSTER --enable-oidc-issuer --enable-workload-identity --enable-managed-identity --generate-ssh-keys
```

Connect to the AKS cluster.

```bash
az aks get-credentials --resource-group $RESOURCE_GROUP --name $MY_CLUSTER
```

If you do not have `kubectl` installed locally, you can install using the following Azure CLI command.

```bash
az aks install-cli
```

#### Enable Workload Identity and OIDC Issuer features
The *gpu-provisioner* controller requires the [workload identity](https://learn.microsoft.com/azure/aks/workload-identity-overview?tabs=dotnet) feature to acquire the access token to the AKS cluster. 

> Run the following commands only if your AKS cluster does not already have the Workload Identity and OIDC issuer features enabled.

```bash
export RESOURCE_GROUP="myResourceGroup"
export MY_CLUSTER="myCluster"
az aks update -g $RESOURCE_GROUP -n $MY_CLUSTER --enable-oidc-issuer --enable-workload-identity --enable-managed-identity
```

#### Create an identity and assign permissions
The identity `kaitoprovisioner` is created for the *gpu-provisioner* controller. It is assigned Contributor role for the managed cluster resource to allow changing `$MY_CLUSTER` (e.g., provisioning new nodes in it).
```bash
export SUBSCRIPTION=$(az account show --query id -o tsv)
export IDENTITY_NAME="kaitoprovisioner"
az identity create --name $IDENTITY_NAME -g $RESOURCE_GROUP
export IDENTITY_PRINCIPAL_ID=$(az identity show --name $IDENTITY_NAME -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query 'principalId' -o tsv)
export IDENTITY_CLIENT_ID=$(az identity show --name $IDENTITY_NAME -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query 'clientId' -o tsv)
az role assignment create --assignee $IDENTITY_PRINCIPAL_ID --scope /subscriptions/$SUBSCRIPTION/resourceGroups/$RESOURCE_GROUP/providers/Microsoft.ContainerService/managedClusters/$MY_CLUSTER  --role "Contributor"
```

#### Install helm charts
Two charts will be installed in `$MY_CLUSTER`: `gpu-provisioner` chart and `workspace` chart.

> Be sure you've cloned this repo and connected to your AKS cluster before attempting to install the Helm charts.

Install the Workspace controller.

```bash
helm install workspace ./charts/kaito/workspace
```

Install the Node provisioner controller.
```bash
# get additional values for helm chart install
export NODE_RESOURCE_GROUP=$(az aks show -n $MY_CLUSTER -g $RESOURCE_GROUP --query nodeResourceGroup -o tsv)
export LOCATION=$(az aks show -n $MY_CLUSTER -g $RESOURCE_GROUP --query location -o tsv)
export TENANT_ID=$(az account show --query tenantId -o tsv)

# create a local values override file
cat << EOF > values.override.yaml
controller:
  env:
  - name: ARM_SUBSCRIPTION_ID
    value: $SUBSCRIPTION
  - name: LOCATION
    value: $LOCATION
  - name: AZURE_CLUSTER_NAME
    value: $MY_CLUSTER
  - name: AZURE_NODE_RESOURCE_GROUP
    value: $NODE_RESOURCE_GROUP
  - name: ARM_RESOURCE_GROUP
    value: $RESOURCE_GROUP
  - name: LEADER_ELECT
    value: "false"
workloadIdentity:
  clientId: $IDENTITY_CLIENT_ID
  tenantId: $TENANT_ID
settings:
  azure:
    clusterName: $MY_CLUSTER
EOF

# install gpu-provisioner using values override file
helm install gpu-provisioner ./charts/kaito/gpu-provisioner -f values.override.yaml
```

#### Create the federated credential
The federated identity credential between the managed identity `kaitoprovisioner` and the service account used by the *gpu-provisioner* controller is created.
```bash
export AKS_OIDC_ISSUER=$(az aks show -n $MY_CLUSTER -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query "oidcIssuerProfile.issuerUrl" -o tsv)
az identity federated-credential create --name kaito-federatedcredential --identity-name $IDENTITY_NAME -g $RESOURCE_GROUP --issuer $AKS_OIDC_ISSUER --subject system:serviceaccount:"gpu-provisioner:gpu-provisioner" --audience api://AzureADTokenExchange --subscription $SUBSCRIPTION
```
Then the *gpu-provisioner* can access the managed cluster using a trust token with the same permissions of the `kaitoprovisioner` identity.
Note that before finishing this step, the *gpu-provisioner* controller pod will constantly fail with the following message in the log:
```
panic: Configure azure client fails. Please ensure federatedcredential has been created for identity XXXX.
```
The pod will reach running state once the federated credential is created.

#### Verify installation
You can run the following commands to verify the installation of the controllers were successful.

Check status of the Helm chart installations.

```bash
helm list -n default
```

Check status of the `workspace`.

```bash
kubectl describe deploy workspace -n workspace
```

Check status of the `gpu-provisioner`.

```bash
kubectl describe deploy gpu-provisioner -n gpu-provisioner
```

#### Troubleshooting 
If you see that the `gpu-provisioner` deployment is not running after some time, it's possible that some values incorrect in your `values.ovveride.yaml`. 

Run the following command to check `gpu-provisioner` pod logs for additional details.

```bash
kubectl logs --selector=app.kubernetes.io\/name=gpu-provisioner -n gpu-provisioner
```

#### Clean up

```bash
helm uninstall gpu-provisioner
helm uninstall workspace
```

## Quick start
After installing Kaito, one can try following commands to start a falcon-7b inference service.
```
$ cat examples/kaito_workspace_falcon_7b.yaml
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

$ kubectl apply -f examples/kaito_workspace_falcon_7b.yaml
```

The workspace status can be tracked by running the following command. When the WORKSPACEREADY column becomes `True`, the model has been deployed successfully.  
```
$ kubectl get workspace workspace-falcon-7b
NAME                  INSTANCE            RESOURCEREADY   INFERENCEREADY   WORKSPACEREADY   AGE
workspace-falcon-7b   Standard_NC12s_v3   True            True             True             10m
```

Next, one can find the inference service's cluster ip and use a temporal `curl` pod to test the service endpoint in the cluster.
```
$ kubectl get svc workspace-falcon-7b
NAME                  TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)            AGE
workspace-falcon-7b   ClusterIP   <CLUSTERIP>  <none>        80/TCP,29500/TCP   10m

export CLUSTERIP=$(kubectl get svc workspace-falcon-7b -o jsonpath="{.spec.clusterIPs[0]}") 
$ kubectl run -it --rm --restart=Never curl --image=curlimages/curl -- curl -X POST http://$CLUSTERIP/chat -H "accept: application/json" -H "Content-Type: application/json" -d "{\"prompt\":\"YOUR QUESTION HERE\"}"
```

## Usage

The detailed usage for Kaito supported models can be found in [**HERE**](presets/README.md). In case users want to deploy their own containerized models, they can provide the pod template in the `inference` field of the workspace custom resource (please see [API definitions](api/v1alpha1/workspace_types.go) for details). The controller will create a deployment workload using all provisioned GPU nodes. Note that currently the controller does **NOT** handle automatic model upgrade. It only creates inference workloads based on the preset configurations if the workloads do not exist.

## Contributing

[Read more](docs/contributing/readme.md)
<!-- markdown-link-check-disable -->
This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit <https://cla.opensource.microsoft.com>.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks
This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow [Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.

## License

See [LICENSE](LICENSE).

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

<!-- markdown-link-check-enable -->
## Contact

"Kaito devs" <kaito@microsoft.com>



