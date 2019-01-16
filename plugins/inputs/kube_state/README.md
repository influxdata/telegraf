# Kube_State Plugin
[kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) is an open source project designed to generate metrics derived from the state of Kubernetes objects – the abstractions Kubernetes uses to represent your cluster. This plugin collects metrics in a similar manner for the following Kubernetes resources:
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
  url = "https://127.0.0.1"

  ## Namespace to use
  # namespace = "default"

  ## Use bearer token for authorization. ('bearer_token' takes priority)
  # bearer_token = "/path/to/bearer/token"
  ## OR
  # bearer_token_string = "abc_123"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional Resources to exclude from gathering
  ## Leave them with blank with try to gather everything available.
  ## Values can be - "configmaps", "daemonsets", deployments", "nodes",
  ## "persistentvolumes", "persistentvolumeclaims", "pods", "statefulsets"
  # resource_exclude = [ "deployments", "nodes", "statefulsets" ]

  ## Optional Resources to include when gathering
  ## Overrides resource_exclude if both set.
  # resource_include = [ "deployments", "nodes", "statefulsets" ]

  ## Optional max age for config map
  # max_config_map_age = "1h"

  ## Optional TLS Config
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

#### Kubernetes Permissions

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


### Metrics:

- kubernetes_configmap
  - tags:
    - configmap_name
    - namespace
    - resource_version
  - fields:
    - created

+ kubernetes_daemonset
  - tags:
    - daemonset_name
    - namespace
  - fields:
    - generation
    - current_number_scheduled
    - desired_number_scheduled
    - number_available
    - number_misscheduled
    - number_ready
    - number_unavailable
    - updated_number_scheduled

- kubernetes_deployment
  - tags:
    - deployment_name
    - namespace
  - fields:
    - replicas_available
    - replicas_unavailable
    - created

+ kubernetes_node
  - tags:
    - node_name
  - fields:
    - capacity_cpu_cores
    - capacity_memory_bytes
    - capacity_pods
    - allocatable_cpu_cores
    - allocatable_memory_bytes
    - allocatable_pods

- kubernetes_persistentvolume
  - tags:
    - pv_name
    - phase
    - storageclass
  - fields:
    - phase_type

+ kubernetes_persistentvolumeclaim
  - tags:
    - pvc_name
    - namespace
    - phase
    - storageclass
  - fields:
    - phase_type

- kubernetes_pod_container
  - tags:
    - container_name
    - namespace
    - node_name
    - pod_name
  - fields:
    - restarts_total
    - running
    - terminated
    - terminated_reason
    - resource_requests_cpu_units
    - resource_requests_memory_bytes
    - resource_limits_cpu_units
    - resource_limits_memory_bytes

+ kubernetes_pod_status
  - tags:
    - namespace
    - pod_name
    - node_name
    - reason
  - fields:
    - last_transition_time
    - ready

- kubernetes_statefulset
  - tags:
    - statefulset_name
    - namespace
  - fields:
    - created
    - generation
    - replicas
    - replicas_current
    - replicas_ready
    - replicas_updated
    - spec_replicas
    - observed_generation


### Example Output:

```
kubernetes_configmap,configmap_name=envoy-config,namespace=default,resource_version=56593031 created=1544103867000000000i 1547597616000000000
kubernetes_daemonset
kubernetes_deployment,deployment_name=deployd,namespace=default replicas_unavailable=0i,created=1544103082000000000i,replicas_available=1i 1547597616000000000
kubernetes_node,node_name=ip-172-17-0-2.internal allocatable_pods=110i,capacity_memory_bytes="125817904Ki",capacity_pods=110i,capacity_cpu_cores=16i,allocatable_cpu_cores=16i,allocatable_memory_bytes="125715504Ki" 1547597616000000000
kubernetes_persistentvolume,phase=Released,pv_name=pvc-aaaaaaaa-bbbb-cccc-1111-222222222222,storageclass=ebs-1-retain phase_type=3i 1547597616000000000
kubernetes_persistentvolumeclaim,namespace=default,phase=Bound,pvc_name=data-etcd-0,storageclass=ebs-1-retain phase_type=0i 1547597615000000000
kubernetes_pod,namespace=default,node_name=ip-172-17-0-2.internal,pod_name=tick1 last_transition_time=1547578322000000000i,ready="false" 1547597616000000000
kubernetes_pod_container,container_name=telegraf,namespace=default,node_name=ip-172-17-0-2.internal,pod_name=tick1 restarts_total=0i,running=1i,terminated=0i,terminated_reason="" 1547597616000000000
kubernetes_statefulset,namespace=default,statefulset_name=etcd replicas_updated=3i,spec_replicas=3i,observed_generation=1i,created=1544101669000000000i,generation=1i,replicas=3i,replicas_current=3i,replicas_ready=3i 1547597616000000000
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
