# KAITO Helm Chart

## Install

```bash
export REGISTRY=<your_docker_registry>
export IMG_NAME=kaito
export IMG_TAG=0.1.0
helm install kaito-workspace ./charts/kaito  --set image.repository=${REGISTRY}/$(IMG_NAME) --set image.tag=$(IMG_TAG)
```

## Configuration 

The following table lists the configurable parameters of the KAITO chart and their default values.

| Parameter                                  | Description | Default                   |
|--------------------------------------------|-------------|-------------------------- |
| `replicaCount`                             |             | `1`                       |
| `image.repository`                         |             | `ghcr.io/Azure/kaito/kaito`   |
| `image.pullPolicy`                         |             | `"IfNotPresent"`          |
| `image.tag`                                |             | `latest`                  |
| `imagePullSecrets`                         |             | `[]`                      |
| `podAnnotations`                           |             | `{}`                      |
| `podSecurityContext.runAsNonRoot`          |             | `true`                    |
| `securityContext.allowPrivilegeEscalation` |             | `false`                   |
| `securityContext.capabilities.drop`        |             | `["ALL"]`                 |
| `resources.limits.cpu`                     |             | `"500m"`                  |
| `resources.limits.memory`                  |             | `"128Mi"`                 |
| `resources.requests.cpu`                   |             | `"10m"`                   |
| `resources.requests.memory`                |             | `"64Mi"`                  |
| `nodeSelector`                             |             | `{}`                      |
| `tolerations`                              |             | `[]`                      |
| `affinity`                                 |             | `{}`                      |
