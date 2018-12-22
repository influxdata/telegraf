# Kubernetes Input Plugin

This input plugin talks to the kubelet api using the `/stats/summary` endpoint to gather metrics about the running pods and containers for a single host. It is assumed that this plugin is running as part of a `daemonset` within a kubernetes installation. This means that telegraf is running on every node within the cluster. Therefore, you should configure this plugin to talk to its locally running kubelet.

To find the ip address of the host you are running on you can issue a command like the following:
```
$ curl -s $API_URL/api/v1/namespaces/$POD_NAMESPACE/pods/$HOSTNAME --header "Authorization: Bearer $TOKEN" --insecure | jq -r '.status.hostIP'
```
In this case we used the downward API to pass in the `$POD_NAMESPACE` and `$HOSTNAME` is the hostname of the pod which is set by the kubernetes API.

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

### Configuration

```toml
[[inputs.kubernetes]]
  ## URL for the kubelet
  url = "http://127.0.0.1:10255"

  ## Use bearer token for authorization
  # bearer_token = /path/to/bearer/token

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### DaemonSet

For recommendations on running Telegraf as a DaemonSet see [Monitoring Kubernetes
Architecture][k8s-telegraf] or view the [Helm charts][tick-charts].

### Metrics

- kubernetes_node
  - tags:
    - node_name
  - fields:
	- cpu_usage_nanocores
	- cpu_usage_core_nanoseconds
	- memory_available_bytes
	- memory_usage_bytes
	- memory_working_set_bytes
	- memory_rss_bytes
	- memory_page_faults
	- memory_major_page_faults
	- network_rx_bytes
	- network_rx_errors
	- network_tx_bytes
	- network_tx_errors
	- fs_available_bytes
	- fs_capacity_bytes
	- fs_used_bytes
	- runtime_image_fs_available_bytes
	- runtime_image_fs_capacity_bytes
	- runtime_image_fs_used_bytes

- kubernetes_pod_container
  - tags:
    - container_name
    - namespace
    - node_name
    - pod_name
  - fields:
	- cpu_usage_nanocores
	- cpu_usage_core_nanoseconds
	- memory_usage_bytes
	- memory_working_set_bytes
	- memory_rss_bytes
	- memory_page_faults
	- memory_major_page_faults
	- rootfs_available_bytes
	- rootfs_capacity_bytes
	- rootfs_used_bytes
	- logsfs_avaialble_bytes
	- logsfs_capacity_bytes
	- logsfs_used_bytes

- kubernetes_pod_volume
  - tags:
    - volume_name
    - namespace
    - node_name
    - pod_name
  - fields:
    - available_bytes
    - capacity_bytes
    - used_bytes

- kubernetes_pod_network
  - tags:
    - namespace
    - node_name
    - pod_name
  - fields:
    - rx_bytes
    - rx_errors
    - tx_bytes
    - tx_errors

### Example Output

```
kubernetes_pod_container,host=ip-10-0-0-0.ec2.internal,container_name=deis-controller,namespace=deis,node_name=ip-10-0-0-0.ec2.internal,pod_name=deis-controller-3058870187-xazsr cpu_usage_core_nanoseconds=2432835i,cpu_usage_nanocores=0i,logsfs_avaialble_bytes=121128271872i,logsfs_capacity_bytes=153567944704i,logsfs_used_bytes=20787200i,memory_major_page_faults=0i,memory_page_faults=175i,memory_rss_bytes=0i,memory_usage_bytes=0i,memory_working_set_bytes=0i,rootfs_available_bytes=121128271872i,rootfs_capacity_bytes=153567944704i,rootfs_used_bytes=1110016i 1476477530000000000
kubernetes_pod_volume,host=ip-10-0-0-0.ec2.internal,name=default-token-f7wts,namespace=kube-system,node_name=ip-10-0-0-0.ec2.internal,pod_name=kubernetes-dashboard-v1.1.1-t4x4t available_bytes=8415240192i,capacity_bytes=8415252480i,used_bytes=12288i 1476477530000000000
kubernetes_pod_network,host=ip-10-0-0-0.ec2.internal,namespace=deis,node_name=ip-10-0-0-0.ec2.internal,pod_name=deis-controller-3058870187-xazsr rx_bytes=120671099i,rx_errors=0i,tx_bytes=102451983i,tx_errors=0i 1476477530000000000
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
