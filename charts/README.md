# KDM Helm Chart

## Install

```bash
helm install kdm ./charts/kdm  --set image.repository=${REGISTRY}/$(IMG_NAME)
```

## Configuration 

The following table lists the configurable parameters of the KDM chart and their default values.

| Parameter                                  | Description | Default          |
|--------------------------------------------|-------------|------------------|
| `replicaCount`                             |             | `1`              |
| `image.repository`                         |             | `"helayoty/kdm"` |
| `image.pullPolicy`                         |             | `"IfNotPresent"` |
| `image.tag`                                |             | `"0.1.0"`        |
| `imagePullSecrets`                         |             | `[]`             |
| `podAnnotations`                           |             | `{}`             |
| `podSecurityContext.runAsNonRoot`          |             | `true`           |
| `securityContext.allowPrivilegeEscalation` |             | `false`          |
| `securityContext.capabilities.drop`        |             | `["ALL"]`        |
| `resources.limits.cpu`                     |             | `"500m"`         |
| `resources.limits.memory`                  |             | `"128Mi"`        |
| `resources.requests.cpu`                   |             | `"10m"`          |
| `resources.requests.memory`                |             | `"64Mi"`         |
| `nodeSelector`                             |             | `{}`             |
| `tolerations`                              |             | `[]`             |
| `affinity`                                 |             | `{}`             |
