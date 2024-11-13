# Use Kaito in AKS Arc
This article describes how to deploy AI models on AKS arc with AI toolchain operator (KAITO). The AI toolchain operator (KAITO) is a managed add-on for all AKS, and it simplifies the experience of running OSS AI models on your AKS clusters. You may follow the workflow below to enable this feature:
1.	Create a Node pool with GPU
2.	Deploy KAITO operator
3.	Deploy AI model
4.	Validate the deployment

## Supported Models
Currently KAITO supports models such as Falcon, Phi2, Phi3, Llama2, Llama2Chat, Mistral. Please refer to KAITO’s [readme](https://github.com/Azure/kaito/blob/main/presets/README.md) file for the latest models. 

## Prerequisite
1.	Before you begin, please make sure you have the following details from your infrastructure administrator:
    - An AKS cluster that's up and running.
    - We recommend using Linux machine for this feature.
    - Your local kubectl environment configured to point to your AKS cluster.
        - Run `az aksarc get-credentials --resource-group <ResourceGroupName> --name <ClusterName>  --admin` to download the kubeconfig file.
2.	Make sure your HCI cluster is enabled with GPU, you can ask your infrastructure administrator to set it up for you. You also need to identify the right VM SKUs for your AKS cluster before creating the node pool. The instruction can be found at [use GPU for compute-intensive workloads](https://learn.microsoft.com/en-us/azure/aks/hybrid/deploy-gpu-node-pool).
3.	Make sure the helm and kubectl are installed in your local machine.
    - If you need to install or upgrade, please see instruction from [Install Helm](https://helm.sh/docs/intro/install/).
    - If you need to install kubectl, please see instructions from [Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

## Create a GPU Node Pool
<details>
<summary><b>Using Azure Portal</b></summary>
<div align="middle">
  <img src="img/aksarc_nodepool_creation_portal.png" width=80% title="create nodepool from azure portal" alt="create nodepool from azure portal">
</div>
</details>

Run following Az command to provision node pool, available GPU sku can be found [here](https://learn.microsoft.com/en-us/azure/aks/hybrid/deploy-gpu-node-pool#supported-vm-sizes)

```bash
az aksarc nodepool add --name "samplenodepool" --cluster-name "samplecluster" --resource-group "sample-rg" –node-vm-size "samplenodepoolsize" –os-type "Linux"
```

### Validation
1.	After node pool creation command succeeds, you can confirm whether the GPU node is provisioned using `kubectl get nodes`.
```
NAME              STATUS   ROLES           AGE   VERSION
moc-l06l9ruvcd6   Ready    <none>          58d   v1.29.4
moc-l9f0lh9ro95   Ready    control-plane   58d   v1.29.4
moc-le4aoguwyd9   Ready    <none>          49d   v1.29.4
```
2.	Please also ensure the node has allocatable GPU cores using command 
```bash
kubectl get node moc-l1i9uh0ksne -o yaml | grep -A 10 "allocatable:"
```
```
  allocatable:
    cpu: 31580m
    ephemeral-storage: "95026644016"
    hugepages-1Gi: "0"
    hugepages-2Mi: "0"
    memory: 121761176Ki
    nvidia.com/gpu: "2"
    pods: "110"
  capacity:
    cpu: "32"
    ephemeral-storage: 103110508Ki
```

## Deploy KAITO via Helm
1.	Git clone the KAITO repo to your local machine
2.	Install KAITO operator using command 
```bash
helm install workspace ./charts/kaito/workspace --namespace kaito-workspace --create-namespace
```

## Deploy LLM Model
<details>
<summary><b>Explain the Yaml file</b></summary>
If a user runs Kaito in an on-premise Kubernetes cluster where nodepool auto provision are unavailable, the GPU nodes can be pre-configured.

- the user needs to add the node names in the `preferredNodes` field in the `resource` spec. As a result, the Kaito controller will skip the steps for GPU node provisioning and use the prepared nodes to run the inference workload.

Using the same method user can try all kaito functionalities, example can be found on /examples folder.
</details>

1.	Create a YAML file with the following template, make sure to replace the placeholders in curly braces with your own information. Please note, the PresetName can be found from the supported model file in KAITO’s github repo.
```yaml
apiVersion: kaito.sh/v1alpha1
kind: Workspace
metadata:
  name: { YourDeploymentName }
resource:
  instanceType: Standard_NC12s_v3
  labelSelector:
    matchLabels:
      apps: { YourNodeLabel }
  preferredNodes:
  - { YourNodeName }
inference:
  preset:
    name: { PresetName }

```
a sample yaml file can be 
```yaml
apiVersion: kaito.sh/v1alpha1
kind: Workspace
metadata:
  name: workspace-falcon-7b
resource:
  instanceType: Standard_NC12s_v3
  labelSelector:
    matchLabels:
      apps: falcon-7b
  preferredNodes: 
  - moc-lmkq7webq9z
inference:
  preset:
    name: falcon-7b-instruct
```
2.	You need to label your GPU node first, `Kubectl label node samplenode app=YourNodeLabel` and then apply the YAML file
`kubectl apply -f sampleyamlfile.yaml`

 

## Validate model deployment 
1.	Validate the workspace using the command `kubectl get workspace`. Please also make sure both `ResourceReady` and `InferenceReady` fields are True before testing with the sample prompt.
```
NAME                           INSTANCE           RESOURCEREADY   INFERENCEREADY   JOBSTARTED   WORKSPACESUCCEEDED   AGE
workspace-falcon-7b            Standard_NC6s_v3   True            True                          True                 4d
```

2.	You may test the model with a sample prompt: 
```bash
export CLUSTERIP=$(kubectl get svc workspace-falcon-7b -o jsonpath="{.spec.clusterIPs[0]}") 

kubectl run -it --rm --restart=Never curl --image=curlimages/curl -- curl -X POST http://$CLUSTERIP/chat -H "accept: application/json" -H "Content-Type: application/json" -d "{\"prompt\":\"<sample_prompt>\"}"
```
a sample output will be like
```bash
$ kubectl run -it --rm --restart=Never curl --image=curlimages/curl -- curl -X POST http://$CLUSTERIP/chat -H "accept: application/json" -H "Content-Type: application/json" -d "{\"prompt\":\"write a poem\"}"
If you don't see a command prompt, try pressing enter.
{"Result":"write a poem about the first day of school, the last day of school, or a day in school.\nThe first day of school\nI wake at dawn\nand think of the new day\nas the sun rises\nI am excited\nthe first day of school\nI wake at dawn\nand think of the new day\nas the sun rises\nI am excited\nthe last day of school\nI wake at dawn\nand think of the last day\nas the sun rises\nI am sad\nthe last day of school\nI wake at dawn\nand think of the last day\nas the sun rises\nI am sad\na day in school\nthe first day of school\nI wake at dawn\nand think of the new day\nas the sun rises\nI am excited\nthe last day of school\nI wake at dawn\nand think of the last day\nas the sun rises\nI am sad\na day in school\nI walk down the stairs"}pod "curl" deleted
```