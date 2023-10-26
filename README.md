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
az aks update -g $RESOURCE_GROUP -n myAKSCluster --enable-oidc-issuer --enable-workload-identity --enable-managed-identity
```

### Create an identity and assign it Contributor role for cluster's resource group
This identity `kaitoprovisioner` is created dedicatedly for the `gpu-povisioner`.
```bash
export SUBSCRIPTION="mySubscription"
az identity create --name kaitoprovisioner -g $RESOURCE_GROUP
export IDENTITY_PRINCIPAL_ID=$(az identity show --name kaitoprovisioner -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query 'principalId' | tr -d '"')
export IDENTITY_CLIENT_ID=$(az identity show --name kaitoprovisioner -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query 'clientId' | tr -d '"')
az role assignment create --assignee $IDENTITY_PRINCIPAL_ID --scope /subscriptions/$SUBSCRIPTION/resourceGroups/$RESOURCE_GROUP  --role "Contributor"

```

### Install helm chart
Two charts will be installed in `myAKSCluster`. One for gpu-provisioner controller, another for workspace controller.
```bash
helm install workspace ./charts/kaito/workspace

export NODE_RESOURCE_GROUP=$(az aks show -n myAKSCluster -g $RESOURCE_GROUP --query nodeResourceGroup | tr -d '"')
export LOCATION=$(az aks show -n myAKSCluster -g $RESOURCE_GROUP --query location | tr -d '"')
export TENANT_ID=$(az account show | jq -r ".tenantId")
yq -i '(.controller.env[] | select(.name=="ARM_SUBSCRIPTION_ID"))       .value = env(SUBSCRIPTION_ID)     ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.controller.env[] | select(.name=="LOCATION"))                  .value = env(LOCATION)            ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.controller.env[] | select(.name=="ARM_RESOURCE_GROUP"))        .value = env(RESOURCE_GROUP)      ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.controller.env[] | select(.name=="AZURE_NODE_RESOURCE_GROUP")) .value = env(NODE_RESOURCE_GROUP) ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.controller.env[] | select(.name=="AZURE_CLUSTER_NAME"))        .value = myAKSCluster             ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.workloadIdentity.clientId)                                            = env(IDENTITY_CLIENT_ID)  ./charts/kaito/gpu-provisioner/values.yaml
yq -i '(.workloadIdentity.tenantId)                                            = env(TENANT_ID)           ./charts/kaito/gpu-provisioner/values.yaml
helm install gpu-provisioner ./charts/kaito/gpu-provisioner 

```

### Create federated credential for the `gpu-provisioner` controller
Allow `gpu-provisioner` controller to use `kaitoprovisioner` identity to operate `myAKSCluster` (e.g., provisioning new nodes) which has been granted sufficient permissions.
```bash
export AKS_OIDC_ISSUER=$(az aks show -n myAKSCluster -g $RESOURCE_GROUP --subscription $SUBSCRIPTION --query "oidcIssuerProfile.issuerUrl" | tr -d '"')
az identity federated-credential create --name kaito-federatedcredential --identity-name kaitoprovisioner -g $RESOURCE_GROUP --issuer $AKS_OIDC_ISSUER --subject system:serviceaccount:"kaito:gpu-provisioner" --audience api://AzureADTokenExchange --subscription $SUBSCRIPTION

```


<details>
<summary>Workspace status</summary>

```bash
Name:         workspace-llama-2-7b-aks
Annotations:  kubernetes-kaito.sh/service-type: load-balancer
API Version:  kaito.sh/v1alpha1
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

### Clean up

```bash
helm uninstall gpu-provisioner
helm uninstall workspace

```



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



