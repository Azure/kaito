# Kubernetes AI Toolchain Operator (Kaito)

[![Go Report Card](https://goreportcard.com/badge/github.com/Azure/kaito)](https://goreportcard.com/report/github.com/Azure/kaito)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Azure/kaito)

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
The following guidance assumes **Azure Kubernetes Service(AKS)** is used to host the Kubernetes cluster .

#### Enable Workload Identity and OIDC Issuer features
The *gpu-povisioner* controller requires the [workload identity](https://learn.microsoft.com/en-us/azure/aks/workload-identity-overview?tabs=dotnet) feature to acquire the access token to the AKS cluster.

```bash
export RESOURCE_GROUP="myResourceGroup"
export MY_CLUSTER="myCluster"
az aks update -g $RESOURCE_GROUP -n $MY_CLUSTER --enable-oidc-issuer --enable-workload-identity --enable-managed-identity
```

#### Create an identity and assign permissions
The identity `kaitoprovisioner` is created for the *gpu-povisioner* controller. It is assigned Contributor role for the managed cluster resource to allow changing `$MY_CLUSTER` (e.g., provisioning new nodes in it).
```bash
export SUBSCRIPTION="mySubscription"
az identity create --name kaitoprovisioner -g $RESOURCE_GROUP
export IDENTITY_PRINCIPAL_ID=$(az identity show --name kaitoprovisioner -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query 'principalId' | tr -d '"')
export IDENTITY_CLIENT_ID=$(az identity show --name kaitoprovisioner -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query 'clientId' | tr -d '"')
az role assignment create --assignee $IDENTITY_PRINCIPAL_ID --scope /subscriptions/$SUBSCRIPTION/resourceGroups/$RESOURCE_GROUP/providers/Microsoft.ContainerService/managedClusters/$MY_CLUSTER  --role "Contributor"

```

#### Install helm charts
Two charts will be installed in `$MY_CLUSTER`: `gpu-provisioner` chart and `workspace` chart.
```bash
helm install workspace ./charts/kaito/workspace

export NODE_RESOURCE_GROUP=$(az aks show -n $MY_CLUSTER -g $RESOURCE_GROUP --query nodeResourceGroup | tr -d '"')
export LOCATION=$(az aks show -n $MY_CLUSTER -g $RESOURCE_GROUP --query location | tr -d '"')
export TENANT_ID=$(az account show | jq -r ".tenantId")
yq -i '(.controller.env[] | select(.name=="ARM_SUBSCRIPTION_ID"))       .value = env(SUBSCRIPTION)'        ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.controller.env[] | select(.name=="LOCATION"))                  .value = env(LOCATION)'            ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.controller.env[] | select(.name=="ARM_RESOURCE_GROUP"))        .value = env(RESOURCE_GROUP)'      ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.controller.env[] | select(.name=="AZURE_NODE_RESOURCE_GROUP")) .value = env(NODE_RESOURCE_GROUP)' ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.controller.env[] | select(.name=="AZURE_CLUSTER_NAME"))        .value = env(MY_CLUSTER)'          ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.settings.azure.clusterName)                                           = env(MY_CLUSTER)'          ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.workloadIdentity.clientId)                                            = env(IDENTITY_CLIENT_ID)'  ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.workloadIdentity.tenantId)                                            = env(TENANT_ID)'           ./charts/kaito/gpu-provisioner/values.yaml
helm install gpu-provisioner ./charts/kaito/gpu-provisioner 

```

#### Create the federated credential
The federated identity credential between the managed identity `kaitoprovisioner` and the service account used by the *gpu-provisioner* controller is created.
```bash
export AKS_OIDC_ISSUER=$(az aks show -n $MY_CLUSTER -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query "oidcIssuerProfile.issuerUrl" | tr -d '"')
az identity federated-credential create --name kaito-federatedcredential --identity-name kaitoprovisioner -g $RESOURCE_GROUP --issuer $AKS_OIDC_ISSUER --subject system:serviceaccount:"gpu-provisioner:gpu-provisioner" --audience api://AzureADTokenExchange --subscription $SUBSCRIPTION
```
Then the *gpu-provisioner* can access the managed cluster using a trust token with the same permissions of the `kaitoprovisioner` identity.
Note that before finishing this step, the *gpu-provisioner* controller pod will constantly fail with the following message in the log:
```
panic: Configure azure client fails. Please ensure federatedcredential has been created for identity XXXX.
```
The pod will reach running state once the federated credential is created.

#### Clean up

```bash
helm uninstall gpu-provisioner
helm uninstall workspace
```

## Quick start
After installing Kaito, one can try following commands to start a faclon-7b inference service.
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
workspace-falcon-7b   ClusterIP   <CLUSTERIP>           <none>        80/TCP,29500/TCP   10m

$ kubectl run -it --rm --restart=Never curl --image=curlimages/curl sh
~ $ curl -X POST http://<CLUSTERIP>/chat -H "accept: application/json" -H "Content-Type: application/json" -d "{\"prompt\":\"YOUR QUESTION HERE\"}"
```

## Usage

The detailed usage for Kaito supported models can be found in [**HERE**](presets/README.md). In case users want to deploy their own containerized models, they can provide the pod template in the `inference` field of the workspace custom resource (please see [API definitions](api/v1alpha1/workspace_types.go) for details). The controller will create a deployment workload using all provisioned GPU nodes. Note that currently the controller does **NOT** handle automatic model upgrade. It only creates inference workloads based on the preset configurations if the workloads do not exist.

## Contributing

[Read more](docs/contributing/readme.md)

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
<!-- markdown-link-check-disable -->
This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow [Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
<!-- markdown-link-check-enable -->
## License

See [LICENSE](LICENSE).

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Contact

"Kaito devs" <kaito@microsoft.com>



