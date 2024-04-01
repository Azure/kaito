## KAITO InferenceUI Demo

The KAITO InferenceUI Demo provides a sample front-end application that demonstrates
how to interface with the KAITO Workspace for inference tasks. 
This guide covers deploying the front-end as a Helm chart in a Kubernetes environment 
as well as how to run the Python application independently.

### Prerequisites

- Kubernetes cluster with Helm installed (if using helm chart)
- Python environment (if running the Python file directly)
- Access to the KAITO Workspace Service endpoint


## Deploying with Helm
To deploy the KAITO InferenceUI Demo using Helm, set the `workspaceServiceURL` environment variable to your 
Workspace Service  endpoint.

### Configuring the Workspace Service URL
- Using the --set flag:

```
helm install inference-frontend ./charts/DemoUI/inference --set env.workspaceServiceURL="http://<CLUSTER_IP>:80/chat"
```
 - Using a custom `values.override.yaml` file:
   ```
    env:
        workspaceServiceURL: "http://<CLUSTER_IP>:80/chat"
   ```
   Then deploy with custom values file:
    ```
    helm install inference-frontend ./charts/DemoUI/inference -f values.override.yaml
    ```

Replace `<CLUSTER_IP>` with the IP address of your Kubernetes cluster. 
For enhanced reliability, consider using the DNS name of the service within your cluster, 
formatted as `http://<SERVICE_NAME>.<NAMESPACE>.svc.cluster.local:80/chat`

## Running the Python File Directly
1. Setup the environment \
   `pip install chainlit requests`
2. Download the Python file \
   `wget -O inference.py https://raw.githubusercontent.com/Azure/kaito/main/demo/inferenceUI/chainlit.py`
3. Run the application (Replace `<URL_TO_KAITO_WORKSPACE>` with the URL to your Kaito Workspace) \
   `WORKSPACE_SERVICE_URL=<URL_TO_KAITO_WORKSPACE> chainlit run inference.py -w`

---

For additional support or to report issues, please contact the development 
team at <kaito-dev@microsoft.com>.