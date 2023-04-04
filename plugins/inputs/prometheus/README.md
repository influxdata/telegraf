# Prometheus Input Plugin

The prometheus input plugin gathers metrics from HTTP servers exposing metrics
in Prometheus format.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from one or many prometheus clients
[[inputs.prometheus]]
  ## An array of urls to scrape metrics from.
  urls = ["http://localhost:9100/metrics"]

  ## Metric version controls the mapping from Prometheus metrics into Telegraf metrics.
  ## See "Metric Format Configuration" in plugins/inputs/prometheus/README.md for details.
  ## Valid options: 1, 2
  # metric_version = 1

  ## Url tag name (tag containing scrapped url. optional, default is "url")
  # url_tag = "url"

  ## Whether the timestamp of the scraped metrics will be ignored.
  ## If set to true, the gather time will be used.
  # ignore_timestamp = false

  ## An array of Kubernetes services to scrape metrics from.
  # kubernetes_services = ["http://my-service-dns.my-namespace:9100/metrics"]

  ## Kubernetes config file to create client from.
  # kube_config = "/path/to/kubernetes.config"

  ## Scrape Pods
  ## Enable scraping of k8s pods. Further settings as to which pods to scape
  ## are determiend by the 'method' option below. When enabled, the default is
  ## to use annotations to determine whether to scrape or not.
  # monitor_kubernetes_pods = false

  ## Scrape Pods Method
  ## annotations: default, looks for specific pod annotations documented below
  ## settings: only look for pods matching the settings provided, not
  ##   annotations
  ## settings+annotations: looks at pods that match annotations using the user
  ##   defined settings
  # monitor_kubernetes_pods_method = "annotations"

  ## Scrape Pods 'annotations' method options
  ## If set method is set to 'annotations' or 'settings+annotations', these
  ## annotation flags are looked for:
  ## - prometheus.io/scrape: Required to enable scraping for this pod. Can also
  ##     use 'prometheus.io/scrape=false' annotation to opt-out entirely.
  ## - prometheus.io/scheme: If the metrics endpoint is secured then you will
  ##     need to set this to 'https' & most likely set the tls config
  ## - prometheus.io/path: If the metrics path is not /metrics, define it with
  ##     this annotation
  ## - prometheus.io/port: If port is not 9102 use this annotation

  ## Scrape Pods 'settings' method options
  ## When using 'settings' or 'settings+annotations', the default values for
  ## annotations can be modified using with the following options:
  # monitor_kubernetes_pods_scheme = "http"
  # monitor_kubernetes_pods_port = "9102"
  # monitor_kubernetes_pods_path = "/metrics"

  ## Get the list of pods to scrape with either the scope of
  ## - cluster: the kubernetes watch api (default, no need to specify)
  ## - node: the local cadvisor api; for scalability. Note that the config node_ip or the environment variable NODE_IP must be set to the host IP.
  # pod_scrape_scope = "cluster"

  ## Only for node scrape scope: node IP of the node that telegraf is running on.
  ## Either this config or the environment variable NODE_IP must be set.
  # node_ip = "10.180.1.1"

  ## Only for node scrape scope: interval in seconds for how often to get updated pod list for scraping.
  ## Default is 60 seconds.
  # pod_scrape_interval = 60

  ## Restricts Kubernetes monitoring to a single namespace
  ##   ex: monitor_kubernetes_pods_namespace = "default"
  # monitor_kubernetes_pods_namespace = ""
  ## The name of the label for the pod that is being scraped.
  ## Default is 'namespace' but this can conflict with metrics that have the label 'namespace'
  # pod_namespace_label_name = "namespace"
  # label selector to target pods which have the label
  # kubernetes_label_selector = "env=dev,app=nginx"
  # field selector to target pods
  # eg. To scrape pods on a specific node
  # kubernetes_field_selector = "spec.nodeName=$HOSTNAME"

  ## Filter which pod annotations and labels will be added to metric tags
  #
  # pod_annotation_include = ["annotation-key-1"]
  # pod_annotation_exclude = ["exclude-me"]
  # pod_label_include = ["label-key-1"]
  # pod_label_exclude = ["exclude-me"]

  # cache refresh interval to set the interval for re-sync of pods list.
  # Default is 60 minutes.
  # cache_refresh_interval = 60

  ## Scrape Services available in Consul Catalog
  # [inputs.prometheus.consul]
  #   enabled = true
  #   agent = "http://localhost:8500"
  #   query_interval = "5m"

  #   [[inputs.prometheus.consul.query]]
  #     name = "a service name"
  #     tag = "a service tag"
  #     url = 'http://{{if ne .ServiceAddress ""}}{{.ServiceAddress}}{{else}}{{.Address}}{{end}}:{{.ServicePort}}/{{with .ServiceMeta.metrics_path}}{{.}}{{else}}metrics{{end}}'
  #     [inputs.prometheus.consul.query.tags]
  #       host = "{{.Node}}"

  ## Use bearer token for authorization. ('bearer_token' takes priority)
  # bearer_token = "/path/to/bearer/token"
  ## OR
  # bearer_token_string = "abc_123"

  ## HTTP Basic Authentication username and password. ('bearer_token' and
  ## 'bearer_token_string' take priority)
  # username = ""
  # password = ""

  ## Optional custom HTTP headers
  # http_headers = {"X-Special-Header" = "Special-Value"}

  ## Specify timeout duration for slower prometheus clients (default is 5s)
  # timeout = "5s"

  ## deprecated in 1.26; use the timeout option
  # response_timeout = "5s"

  ## HTTP Proxy support
  # use_system_proxy = false
  # http_proxy_url = ""

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile

  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Use the given name as the SNI server name on each URL
  # tls_server_name = "myhost.example.org"

  ## TLS renegotiation method, choose from "never", "once", "freely"
  # tls_renegotiation_method = "never"

  ## Enable/disable TLS
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable = true

  ## Control pod scraping based on pod namespace annotations
  ## Pass and drop here act like tagpass and tagdrop, but instead
  ## of filtering metrics they filters pod candidates for scraping
  #[inputs.prometheus.namespace_annotation_pass]
  # annotation_key = ["value1", "value2"]
  #[inputs.prometheus.namespace_annotation_drop]
  # some_annotation_key = ["dont-scrape"]
```

`urls` can contain a unix socket as well. If a different path is required
(default is `/metrics` for both http[s] and unix) for a unix socket, add `path`
as a query parameter as follows:
`unix:///var/run/prometheus.sock?path=/custom/metrics`

### Metric Format Configuration

The `metric_version` setting controls how telegraf translates prometheus format
metrics to telegraf metrics. There are two options.

With `metric_version = 1`, the prometheus metric name becomes the telegraf
metric name. Prometheus labels become telegraf tags. Prometheus values become
telegraf field values. The fields have generic keys based on the type of the
prometheus metric. This option produces metrics that are dense (not
sparse). Denseness is a useful property for some outputs, including those that
are more efficient with row-oriented data.

`metric_version = 2` differs in a few ways. The prometheus metric name becomes a
telegraf field key. Metrics hold more than one value and the field keys aren't
generic. The resulting metrics are sparse, but for some outputs they may be
easier to process or query, including those that are more efficient with
column-oriented data. The telegraf metric name is the same for all metrics in
the input instance. It can be set with the `name_override` setting and defaults
to "prometheus". To have multiple metric names, you can use multiple instances
of the plugin, each with its own `name_override`.

`metric_version = 2` uses the same histogram format as the [histogram
aggregator](../../aggregators/histogram/README.md)

The Example Outputs sections shows examples for both options.

When using this plugin along with the prometheus_client output, use the same
option in both to ensure metrics are round-tripped without modification.

### Kubernetes Service Discovery

URLs listed in the `kubernetes_services` parameter will be expanded by looking
up all A records assigned to the hostname as described in [Kubernetes DNS
service discovery][serv-disc].

This method can be used to locate all [Kubernetes headless services][headless].

[serv-disc]: https://kubernetes.io/docs/concepts/services-networking/service/#dns

[headless]: https://kubernetes.io/docs/concepts/services-networking/service/#headless-services

### Kubernetes scraping

Enabling this option will allow the plugin to scrape for prometheus annotation
on Kubernetes pods. Currently, you can run this plugin in your kubernetes
cluster, or we use the kubeconfig file to determine where to monitor.  Currently
the following annotation are supported:

* `prometheus.io/scrape` Enable scraping for this pod.
* `prometheus.io/scheme` If the metrics endpoint is secured then you will need to set this to `https` & most likely set the tls config. (default 'http')
* `prometheus.io/path` Override the path for the metrics endpoint on the service. (default '/metrics')
* `prometheus.io/port` Used to override the port. (default 9102)

Using the `monitor_kubernetes_pods_namespace` option allows you to limit which
pods you are scraping.

The setting `pod_namespace_label_name` allows you to change the label name for
the namespace of the pod you are scraping. The default is `namespace`, but this
will overwrite a label with the name `namespace` from a metric scraped.

Using `pod_scrape_scope = "node"` allows more scalable scraping for pods which
will scrape pods only in the node that telegraf is running. It will fetch the
pod list locally from the node's kubelet. This will require running Telegraf in
every node of the cluster. Note that either `node_ip` must be specified in the
config or the environment variable `NODE_IP` must be set to the host IP. ThisThe
latter can be done in the yaml of the pod running telegraf:

```sh
env:
  - name: NODE_IP
    valueFrom:
      fieldRef:
        fieldPath: status.hostIP
 ```

If using node level scrape scope, `pod_scrape_interval` specifies how often (in
seconds) the pod list for scraping should updated. If not specified, the default
is 60 seconds.

The pod running telegraf will need to have the proper rbac configuration in
order to be allowed to call the k8s api to discover and watch pods in the
cluster.  A typical configuration will create a service account, a cluster role
with the appropriate rules and a cluster role binding to tie the cluster role to
the service account.  Example of configuration for cluster level discovery:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: telegraf-k8s-role-{{.Release.Name}}
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
---
# Rolebinding for namespace to cluster-admin
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: telegraf-k8s-role-{{.Release.Name}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: telegraf-k8s-role-{{.Release.Name}}
subjects:
- kind: ServiceAccount
  name: telegraf-k8s-{{ .Release.Name }}
  namespace: {{ .Release.Namespace }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: telegraf-k8s-{{ .Release.Name }}
```

### Consul Service Discovery

Enabling this option and configuring consul `agent` url will allow the plugin to
query consul catalog for available services. Using `query_interval` the plugin
will periodically query the consul catalog for services with `name` and `tag`
and refresh the list of scraped urls.  It can use the information from the
catalog to build the scraped url and additional tags from a template.

Multiple consul queries can be configured, each for different service.
The following example fields can be used in url or tag templates:

* Node
* Address
* NodeMeta
* ServicePort
* ServiceAddress
* ServiceTags
* ServiceMeta

For full list of available fields and their type see struct CatalogService in
<https://github.com/hashicorp/consul/blob/master/api/catalog.go>

### Bearer Token

If set, the file specified by the `bearer_token` parameter will be read on
each interval and its contents will be appended to the Bearer string in the
Authorization header.

## Usage for Caddy HTTP server

Steps to monitor Caddy with Telegraf's Prometheus input plugin:

* Download [Caddy](https://caddyserver.com/download)
* Download Prometheus and set up [monitoring Caddy with Prometheus metrics](https://caddyserver.com/docs/metrics#monitoring-caddy-with-prometheus-metrics)
* Restart Caddy
* Configure Telegraf to fetch metrics on it:

```toml
[[inputs.prometheus]]
#   ## An array of urls to scrape metrics from.
  urls = ["http://localhost:2019/metrics"]
```

> This is the default URL where Caddy will send data.
> For more details, please read the [Caddy Prometheus documentation](https://github.com/miekg/caddy-prometheus/blob/master/README.md).

## Metrics

Measurement names are based on the Metric Family and tags are created for each
label.  The value is added to a field named based on the metric type.

All metrics receive the `url` tag indicating the related URL specified in the
Telegraf configuration. If using Kubernetes service discovery the `address`
tag is also added indicating the discovered ip address.

## Example Output

### Source

```shell
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 7.4545e-05
go_gc_duration_seconds{quantile="0.25"} 7.6999e-05
go_gc_duration_seconds{quantile="0.5"} 0.000277935
go_gc_duration_seconds{quantile="0.75"} 0.000706591
go_gc_duration_seconds{quantile="1"} 0.000706591
go_gc_duration_seconds_sum 0.00113607
go_gc_duration_seconds_count 4
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 15
# HELP cpu_usage_user Telegraf collected metric
# TYPE cpu_usage_user gauge
cpu_usage_user{cpu="cpu0"} 1.4112903225816156
cpu_usage_user{cpu="cpu1"} 0.702106318955865
cpu_usage_user{cpu="cpu2"} 2.0161290322588776
cpu_usage_user{cpu="cpu3"} 1.5045135406226022
```

### Output

```text
go_gc_duration_seconds,url=http://example.org:9273/metrics 1=0.001336611,count=14,sum=0.004527551,0=0.000057965,0.25=0.000083812,0.5=0.000286537,0.75=0.000365303 1505776733000000000
go_goroutines,url=http://example.org:9273/metrics gauge=21 1505776695000000000
cpu_usage_user,cpu=cpu0,url=http://example.org:9273/metrics gauge=1.513622603430151 1505776751000000000
cpu_usage_user,cpu=cpu1,url=http://example.org:9273/metrics gauge=5.829145728641773 1505776751000000000
cpu_usage_user,cpu=cpu2,url=http://example.org:9273/metrics gauge=2.119071644805144 1505776751000000000
cpu_usage_user,cpu=cpu3,url=http://example.org:9273/metrics gauge=1.5228426395944945 1505776751000000000
```

### Output (when metric_version = 2)

```text
prometheus,quantile=1,url=http://example.org:9273/metrics go_gc_duration_seconds=0.005574303 1556075100000000000
prometheus,quantile=0.75,url=http://example.org:9273/metrics go_gc_duration_seconds=0.0001046 1556075100000000000
prometheus,quantile=0.5,url=http://example.org:9273/metrics go_gc_duration_seconds=0.0000719 1556075100000000000
prometheus,quantile=0.25,url=http://example.org:9273/metrics go_gc_duration_seconds=0.0000579 1556075100000000000
prometheus,quantile=0,url=http://example.org:9273/metrics go_gc_duration_seconds=0.0000349 1556075100000000000
prometheus,url=http://example.org:9273/metrics go_gc_duration_seconds_count=324,go_gc_duration_seconds_sum=0.091340353 1556075100000000000
prometheus,url=http://example.org:9273/metrics go_goroutines=15 1556075100000000000
prometheus,cpu=cpu0,url=http://example.org:9273/metrics cpu_usage_user=1.513622603430151 1505776751000000000
prometheus,cpu=cpu1,url=http://example.org:9273/metrics cpu_usage_user=5.829145728641773 1505776751000000000
prometheus,cpu=cpu2,url=http://example.org:9273/metrics cpu_usage_user=2.119071644805144 1505776751000000000
prometheus,cpu=cpu3,url=http://example.org:9273/metrics cpu_usage_user=1.5228426395944945 1505776751000000000
```
