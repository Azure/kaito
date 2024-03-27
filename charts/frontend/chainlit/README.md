# KAITO Sample Frontend (Chainlit) Helm Chart
## Install
Before deploying the Chainlit front-end, you must set the `workspaceServiceURL` environment variable to point to your Workspace Service inference endpoint.

To set this value, modify the `values.override.yaml` file or use the `--set` flag during Helm install/upgrade:

```bash
helm install chainlit-frontend ./charts/frontend/chainlit/values.yaml --set env.workspaceServiceURL="http://<CLUSTER_IP>:80/chat"
```

Or through a custom `values` file (`values.override.yaml`): 
```bash
helm install chainlit-frontend ./charts/frontend/chainlit/values.yaml -f values.override.yaml
```

## Values

| Key                                      | Type   | Default                           | Description |
|------------------------------------------|--------|-----------------------------------|-------------|
| affinity                                 | object | `{}`                              |             |
| image.pullPolicy                         | string | `"IfNotPresent"`                  |             |
| image.repository                         | string | `"python"`                        |             |
| image.tag                                | string | `"3.8"`                           |             |
| imagePullSecrets                         | list   | `[]`                              |             |
| nodeSelector                             | object | `{}`                              |             |
| podAnnotations                           | object | `{}`                              |             |
| replicaCount                             | int    | `1`                               |             |
| resources.limits.cpu                     | string | `"500m"`                          |             |
| resources.limits.memory                  | string | `"128Mi"`                         |             |
| resources.requests.cpu                   | string | `"10m"`                           |             |
| resources.requests.memory                | string | `"64Mi"`                          |             |
| tolerations                              | list   | `[]`                              |             |
