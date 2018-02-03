# Prometheus Input Plugin

The prometheus input plugin gathers metrics from HTTP servers exposing metrics
in Prometheus format.

### Configuration:

```toml
# Read metrics from one or many prometheus clients
[[inputs.prometheus]]
  ## An array of urls to scrape metrics from.
  urls = ["http://localhost:9100/metrics"]

  ## An array of Kubernetes services to scrape metrics from.
  # kubernetes_services = ["http://my-service-dns.my-namespace:9100/metrics"]

  ## Use bearer token for authorization
  # bearer_token = /path/to/bearer/token

  ## Specify timeout duration for slower prometheus clients (default is 3s)
  # response_timeout = "3s"

  ## Optional SSL Config
  # ssl_ca = /path/to/cafile
  # ssl_cert = /path/to/certfile
  # ssl_key = /path/to/keyfile
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
```

#### Kubernetes Service Discovery

URLs listed in the `kubernetes_services` parameter will be expanded
by looking up all A records assigned to the hostname as described in
[Kubernetes DNS service discovery](https://kubernetes.io/docs/concepts/services-networking/service/#dns).

This method can be used to locate all
[Kubernetes headless services](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services).

#### Bearer Token

If set, the file specified by the `bearer_token` parameter will be read on
each interval and its contents will be appended to the Bearer string in the
Authorization header.

### Usage for Caddy HTTP server

If you want to monitor Caddy, you need to use Caddy with its Prometheus plugin:

* Download Caddy+Prometheus plugin [here](https://caddyserver.com/download/linux/amd64?plugins=http.prometheus)
* Add the `prometheus` directive in your `CaddyFile`
* Restart Caddy
* Configure Telegraf to fetch metrics on it:

```
[[inputs.prometheus]]
#   ## An array of urls to scrape metrics from.
  urls = ["http://localhost:9180/metrics"]
```

> This is the default URL where Caddy Prometheus plugin will send data.
> For more details, please read the [Caddy Prometheus documentation](https://github.com/miekg/caddy-prometheus/blob/master/README.md).

### Metrics:

Measurement names are based on the Metric Family and tags are created for each
label.  The value is added to a field named based on the metric type.

All metrics receive the `url` tag indicating the related URL specified in the
Telegraf configuration. If using Kubernetes service discovery the `address`
tag is also added indicating the discovered ip address.

### Example Output:

**Source**
```
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

**Output**
```
go_gc_duration_seconds,url=http://example.org:9273/metrics 1=0.001336611,count=14,sum=0.004527551,0=0.000057965,0.25=0.000083812,0.5=0.000286537,0.75=0.000365303 1505776733000000000
go_goroutines,url=http://example.org:9273/metrics gauge=21 1505776695000000000
cpu_usage_user,cpu=cpu0,url=http://example.org:9273/metrics gauge=1.513622603430151 1505776751000000000
cpu_usage_user,cpu=cpu1,url=http://example.org:9273/metrics gauge=5.829145728641773 1505776751000000000
cpu_usage_user,cpu=cpu2,url=http://example.org:9273/metrics gauge=2.119071644805144 1505776751000000000
cpu_usage_user,cpu=cpu3,url=http://example.org:9273/metrics gauge=1.5228426395944945 1505776751000000000
```
