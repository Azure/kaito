# KAITO Workspace Helm Chart

## Install

```bash
export REGISTRY=mcr.microsoft.com/aks/kaito
export IMG_NAME=ragengine
export IMG_TAG=0.0.1
helm install ragengine ./charts/kaito/ragengine  \
--set image.repository=${REGISTRY}/$(IMG_NAME) --set image.tag=$(IMG_TAG) \
--namespace kaito-ragengine --create-namespace
```

## Values

| Key                                      | Type   | Default                                 | Description                                                   |
|------------------------------------------|--------|-----------------------------------------|---------------------------------------------------------------|
| affinity                                 | object | `{}`                                    |                                                               |
| image.pullPolicy                         | string | `"IfNotPresent"`                        |                                                               |
| image.repository                         | string | `mcr.microsoft.com/aks/kaito/workspace` |                                                               |
| image.tag                                | string | `"0.3.0"`                               |                                                               |
| imagePullSecrets                         | list   | `[]`                                    |                                                               |
| nodeSelector                             | object | `{}`                                    |                                                               |
| podAnnotations                           | object | `{}`                                    |                                                               |
| podSecurityContext.runAsNonRoot          | bool   | `true`                                  |                                                               |
| presetRegistryName                       | string | `"mcr.microsoft.com/aks/kaito"`         |                                                               |
| replicaCount                             | int    | `1`                                     |                                                               |
| resources.limits.cpu                     | string | `"500m"`                                |                                                               |
| resources.limits.memory                  | string | `"128Mi"`                               |                                                               |
| resources.requests.cpu                   | string | `"10m"`                                 |                                                               |
| resources.requests.memory                | string | `"64Mi"`                                |                                                               |
| securityContext.allowPrivilegeEscalation | bool   | `false`                                 |                                                               |
| securityContext.capabilities.drop[0]     | string | `"ALL"`                                 |                                                               |
| tolerations                              | list   | `[]`                                    |                                                               |
| webhook.port                             | int    | `9443`                                  |                                                               |
