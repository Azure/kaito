# KDM Helm Chart

## Install

```bash
export REGISTRY=<your_docker_registry>
export IMG_NAME=kdm
export IMG_TAG=0.1.0
helm install kdm-workspace ./charts/kdm  --set image.repository=${REGISTRY}/$(IMG_NAME) --set image.tag=$(IMG_TAG)
```

## Configuration 

The following table lists the configurable parameters of the KDM chart and their default values.

| Parameter                                  | Description | Default                   |
|--------------------------------------------|-------------|-------------------------- |
| `replicaCount`                             |             | `1`                       |
| `image.repository`                         |             | `ghcr.io/Azure/kdm/kdm`   |
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
