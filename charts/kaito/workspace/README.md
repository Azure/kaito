# KAITO Workspace Helm Chart

## Install

```bash
export REGISTRY=<your_docker_registry>
export IMG_NAME=workspace
export IMG_TAG=0.2.1
helm install workspace ./charts/kaito/workspace  \
--set image.repository=${REGISTRY}/$(IMG_NAME) --set image.tag=$(IMG_TAG) \
--namespace kaito-workspace --create-namespace
```

## Values

| Key                                      | Type   | Default                           | Description |
|------------------------------------------|--------|-----------------------------------|-------------|
| affinity                                 | object | `{}`                              |             |
| image.pullPolicy                         | string | `"IfNotPresent"`                  |             |
| image.repository                         | string | `"ghcr.io/azure/kaito/workspace"` |             |
| image.tag                                | string | `"0.2.0"`                         |             |
| imagePullSecrets                         | list   | `[]`                              |             |
| nodeSelector                             | object | `{}`                              |             |
| podAnnotations                           | object | `{}`                              |             |
| podSecurityContext.runAsNonRoot          | bool   | `true`                            |             |
| presetRegistryName                       | string | `"mcr.microsoft.com/aks/kaito"`   |             |
| replicaCount                             | int    | `1`                               |             |
| resources.limits.cpu                     | string | `"500m"`                          |             |
| resources.limits.memory                  | string | `"128Mi"`                         |             |
| resources.requests.cpu                   | string | `"10m"`                           |             |
| resources.requests.memory                | string | `"64Mi"`                          |             |
| securityContext.allowPrivilegeEscalation | bool   | `false`                           |             |
| securityContext.capabilities.drop[0]     | string | `"ALL"`                           |             |
| tolerations                              | list   | `[]`                              |             |
| webhook.port                             | int    | `9443`                            |             |
