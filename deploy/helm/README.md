# alertmanager-discord

![Version: 0.0.0-local](https://img.shields.io/badge/Version-0.0.0--local-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.0-local](https://img.shields.io/badge/AppVersion-0.0.0--local-informational?style=flat-square)

A Helm chart to deploy alertmanager-discord to Kubernetes

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| ciliumNetworkPolicy.alertManagerSelectorLabels | object | `{}` | the labels applied to the alertmanager which will send data to this service. If Cilium Network Policy is enabled, ingress to this service is only allowed from a pod matching these labels. |
| ciliumNetworkPolicy.enabled | bool | `false` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"Always"` |  |
| image.repository | string | `"speckle/alertmanager-discord"` |  |
| image.tag | string | `"latest"` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext.fsGroup | int | `2000` |  |
| priorityClassName | string | `""` | https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/ |
| replicaCount | int | `1` |  |
| resources.limits.cpu | string | `"100m"` |  |
| resources.limits.memory | string | `"128Mi"` |  |
| resources.requests.cpu | string | `"50m"` |  |
| resources.requests.memory | string | `"64Mi"` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `1000` |  |
| server.configuration.key | string | `"config.yaml"` | the key within the Kubernetes Secret. This key is expected to be a filename, as it will for the path for the configuration file when mounted to the container. |
| server.configuration.name | string | `"discord-config"` | name of the Kubernetes Secret containing the configuration file, will be mounted to the container. Must be in the same namespace as this helm chart is deployed. |
| service.port | int | `9094` | The port to which alertmanager should push alerts |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` |  |

