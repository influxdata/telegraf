# Kube_State Plugin
[kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) is an open source project designed to generate metrics derived from the state of Kubernetes objects – the abstractions Kubernetes uses to represent your cluster. This plugin collects metrics in a similar manner for the following kubernetes resources:
 - configmaps
 - daemonsets
 - deployments
 - nodes
 - persistentvolumes
 - persistentvolumeclaims
 - pods (containers/status)
 - statefulsets

#### Series Cardinality Warning

This plugin may produce a high number of series which, when not controlled
for, will cause high load on your database.  Use the following techniques to
avoid cardinality issues:

- Use [metric filtering][] options to exclude unneeded measurements and tags.
- Write to a database with an appropriate [retention policy][].
- Limit series cardinality in your database using the
  [max-series-per-database][] and [max-values-per-tag][] settings.
- Consider using the [Time Series Index][tsi].
- Monitor your databases [series cardinality][].
- Consult the [InfluxDB documentation][influx-docs] for the most up-to-date techniques.

### Configuration:

```toml
[[inputs.kube_state]]
  ## URL for the kubelet
  url = "https://1.1.1.1"

  ## Namespace to use
  namespace = "default"

  ## Use bearer token for authorization (token has priority over file)
  # bearer_token = abc123
  ## or
  # bearer_token_file = /path/to/bearer/token

  ## Set response_timeout (default 5 seconds)
  #  response_timeout = "5s"

  ## Optional Resources to exclude from gathering
  ## Leave them with blank with try to gather everything available.
  ## Values can be - "configmaps", "daemonsets", deployments", "nodes",
  ## "persistentvolumes", "persistentvolumeclaims", "pods", "statefulsets"
  #  resource_exclude = [ "deployments", "nodes", "statefulsets" ]

  ## Optional Resources to include when gathering
  ## Overrides resource_exclude if both set.
  #  resource_include = [ "deployments", "nodes", "statefulsets" ]

  ## Optional max age for config map
  #  max_config_map_age = "1h"

  ## Optional TLS Config
  ## Use TLS but skip chain & host verification
  #  insecure_skip_verify = false
```


### Metrics:

#### kubernetes_configmap
```
TS: creation time

Fields:
    gauge

Tags:
    configmap_name
    namespace
    resource_version
```

#### kubernetes_daemonset
```
TS: now

Fields:
    metadata_generation
    created
    status_current_number_scheduled
    status_desired_number_scheduled
    status_number_available
    status_number_misscheduled
    status_number_ready
    status_number_unavailable
    status_updated_number_scheduled

Tags:
    daemonset_name
    namespace
```

#### kubernetes_deployment
```
TS: now

Fields:
    status_replicas_available
    status_replicas_unavailable
    created

Tags:
    deployment_name
    namespace
```

#### kubernetes_node
```
TS: now

Fields:
    status_allocatable_cpu_cores
    status_allocatable_memory_bytes
    status_allocatable_pods
    status_capacity_pods
    status_capacity_cpu_cores
    status_capacity_memory_bytes

Tags:
    node_name
```

#### kubernetes_persistentvolume
```
TS: now

Fields:
    status_available
    status_bound
    status_failed
    status_pending
    status_released

Tags:
    pv_name
    storageclass
    status
```

#### kubernetes_persistentvolumeclaim
```
TS: now

Fields:
    status_lost
    status_pending
    status_bound

Tags:
    pvc_name
    namespace
    status
    storageclass
```

#### kubernetes_pod_container
```
TS: now

Fields:
    status_restarts_total
    status_running
    status_terminated
    status_terminated_reason
    resource_requests_cpu_cores
    resource_requests_memory_bytes
    resource_requests_limits_cpu_cores
    resource_requests_limits_memory_bytes

Tags:
    container_name
    namespace
    node_name
    pod_name
```

#### kubernetes_pod_status
```
TS: creation time

Fields:
    ready

Tags:
    pod_name
    namespace
    node_name
    reason  [Completed,ContainerCannotRun,Error,OOMKilled]
```

#### kubernetes_statefulset
```
TS: creation time

Fields:
    metadata_generation
    replicas
    status_replicas
    status_replicas_current
    status_replicas_ready
    status_replicas_updated
    status_observed_generation

Tags:
    ss_name
    namespace
```


### Example Output:

```
kubernetes_configmap,configmap_name=envoy-config,namespace=default,resource_version=56593031 gauge=1i 1544103867000000000
kubernetes_daemonset
kubernetes_deployment,deployment_name=tasks,namespace=default created=1544102512000000000i,status_replicas_available=1i,status_replicas_unavailable=0i 1546915265000000000
kubernetes_node,node_name=ip-172-17-0-1.internal status_capacity_memory_bytes="125817904Ki",status_capacity_pods=110i,status_allocatable_cpu_cores=16i,status_allocatable_memory_bytes="125715504Ki",status_allocatable_pods=110i,status_capacity_cpu_cores=16i 1546978191000000000
kubernetes_persistentvolume,pv_name=pvc-aaaaaaaa-bbbb-cccc-1111-222222222222,status=Bound,storageclass=ebs-1 status_available=0i,status_bound=1i,status_failed=0i,status_pending=0i,status_released=0i 1546978191000000000
kubernetes_persistentvolumeclaim,pvc_name=storage-7,namespace=default,status=Bound,storageclass=ebs-1-retain status_lost=0i,status_bound=1i,status_pending=0i 1546912925000000000
kubernetes_pod_container,container_name=telegraf,namespace=default,node_name=ip-172-17-0-1.internal,pod_name=storage-7 resource_requests_cpu_units="100m",resource_requests_memory_bytes="500Mi",resource_limits_cpu_units="500m",resource_limits_memory_bytes="500Mi",status_restarts_total=1i,status_running=1i,status_terminated=0i,status_terminated_reason="" 1546912926000000000
kubernetes_pod_status,pod_name=storage-7,namespace=default,node_name=ip-172-17-0-2.internal ready="true" 1546910783000000000
kubernetes_statefulset,ss_name=kafka,namespace=default status_replicas=8i,status_replicas_current=8i,status_replicas_ready=8i,status_replicas_updated=8i,replicas=8i,status_observed_generation=4i,metadata_generation=4i 1544101819000000000
```


### Kubernetes Permissions

If using [RBAC authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/), you will need to create a cluster role to list "persistentvolumes" and "nodes". You will then need to make an [aggregated ClusterRole](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) that will eventually be bound to a user or group.
```yaml
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: influx:cluster:viewer
  labels:
    rbac.authorization.k8s.io/aggregate-view-telegraf: "true"
rules:
- apiGroups: [""]
  resources: ["persistentvolumes","nodes"]
  verbs: ["get","list"]

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: influx:telegraf
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.k8s.io/aggregate-view-telegraf: "true"
      rbac.authorization.k8s.io/aggregate-to-view: "true"
rules: [] # Rules are automatically filled in by the controller manager.
```

Bind the newly created aggregated ClusterRole with the following config file, updating the subjects as needed.
```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: influx:telegraf:viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: influx:telegraf
subjects:
- kind: ServiceAccount
  name: telegraf
  namespace: default
```


[metric filtering]: https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#metric-filtering
[retention policy]: https://docs.influxdata.com/influxdb/latest/guides/downsampling_and_retention/
[max-series-per-database]: https://docs.influxdata.com/influxdb/latest/administration/config/#max-series-per-database-1000000
[max-values-per-tag]: https://docs.influxdata.com/influxdb/latest/administration/config/#max-values-per-tag-100000
[tsi]: https://docs.influxdata.com/influxdb/latest/concepts/time-series-index/
[series cardinality]: https://docs.influxdata.com/influxdb/latest/query_language/spec/#show-cardinality
[influx-docs]: https://docs.influxdata.com/influxdb/latest/
[k8s-telegraf]: https://www.influxdata.com/blog/monitoring-kubernetes-architecture/
[tick-charts]: https://github.com/influxdata/tick-charts
