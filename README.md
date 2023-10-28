# Kubernetes AI Toolchain Operator(KAITO)

[![Go Report Card](https://goreportcard.com/badge/github.com/Azure/kaito)](https://goreportcard.com/report/github.com/Azure/kaito)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Azure/kaito)

KAITO has been designed to simplify the workflow of launching AI inference services against popular large open sourced AI models,
such as Falcon or Llama, in a Kubernetes cluster.

## Installation 
The following guidence assumes **Azure Kubernetes Service(AKS)** is used to host the Kubernetes cluster .

### Enable Workload Identity and OIDC Issuer features
The `gpu-povisioner` component requires the [workload identity](https://learn.microsoft.com/en-us/azure/aks/workload-identity-overview?tabs=dotnet) feature to acquire the token to access the AKS managed cluster with proper permissions.

```bash
export RESOURCE_GROUP="myResourceGroup"
export MY_CLUSTER="myCluster"
az aks update -g $RESOURCE_GROUP -n $MY_CLUSTER --enable-oidc-issuer --enable-workload-identity --enable-managed-identity
```

### Create an identity and assign permissions
This identity `kaitoprovisioner` is created for the `gpu-povisioner` controller. It is assigned Contributor role for `$RESOURCE_GROUP` to allow operating `$MY_CLUSTER` (e.g., provisioning new nodes in it).
```bash
export SUBSCRIPTION="mySubscription"
az identity create --name kaitoprovisioner -g $RESOURCE_GROUP
export IDENTITY_PRINCIPAL_ID=$(az identity show --name kaitoprovisioner -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query 'principalId' | tr -d '"')
export IDENTITY_CLIENT_ID=$(az identity show --name kaitoprovisioner -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query 'clientId' | tr -d '"')
az role assignment create --assignee $IDENTITY_PRINCIPAL_ID --scope /subscriptions/$SUBSCRIPTION/resourceGroups/$RESOURCE_GROUP  --role "Contributor"

```

### Install helm charts
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
yq -i '(.workloadIdentity.clientId)                                            = env(IDENTITY_CLIENT_ID)'  ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.workloadIdentity.tenantId)                                            = env(TENANT_ID)'           ./charts/kaito/gpu-provisioner/values.yaml
helm install gpu-provisioner ./charts/kaito/gpu-provisioner 

```

### Create federated credential
This allows `gpu-provisioner` controller to use `kaitoprovisioner` identity via an access token.
```bash
export AKS_OIDC_ISSUER=$(az aks show -n $MY_CLUSTER -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query "oidcIssuerProfile.issuerUrl" | tr -d '"')
az identity federated-credential create --name kaito-federatedcredential --identity-name kaitoprovisioner -g $RESOURCE_GROUP --issuer $AKS_OIDC_ISSUER --subject system:serviceaccount:"gpu-provisioner:gpu-provisioner" --audience api://AzureADTokenExchange --subscription $SUBSCRIPTION
```
Note that before doing this step, the `gpu-provisioner` controller pod will constantly fail with the following message in the log:
```
panic: Configure azure client fails. Please ensure federatedcredential has been created for identity XXXX.
```
The pod will reach running state once the federated credential is created.

### Clean up

```bash
helm uninstall gpu-provisioner
helm uninstall workspace
```

## Quick start

TODO.


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

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow [Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.

## License

See [LICENSE](LICENSE).

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Contact

"Kaito devs" <kaito@microsoft.com>



