# Prometheus Input Plugin

The prometheus input plugin gathers metrics from any webpage
exposing metrics with Prometheus format

### Configuration:

Example for Kubernetes apiserver
```toml
# Get all metrics from Kube-apiserver
[[inputs.prometheus]]
  # An array of urls to scrape metrics from.
  urls = ["http://my-kube-apiserver:8080/metrics"]
```

Specify a 10 second timeout for slower/over-loaded clients
```toml
# Get all metrics from Kube-apiserver
[[inputs.prometheus]]
  # An array of urls to scrape metrics from.
  urls = ["http://my-kube-apiserver:8080/metrics"]

  # Specify timeout duration for slower prometheus clients (default is 3s)
  response_timeout = "10s"
```

You can use more complex configuration
to filter and some tags

```toml
# Get all metrics from Kube-apiserver
[[inputs.prometheus]]
  # An array of urls to scrape metrics from.
  urls = ["http://my-kube-apiserver:8080/metrics"]
  # Get only metrics with "apiserver_" string is in metric name
  namepass = ["apiserver_*"]
  # Add a metric name prefix
  name_prefix = "k8s_"
  # Add tags to be able to make beautiful dashboards
  [inputs.prometheus.tags]
    kubeservice = "kube-apiserver"
```

```toml
# Authorize with a bearer token skipping cert verification
[[inputs.prometheus]]
  # An array of urls to scrape metrics from.
  urls = ["http://my-kube-apiserver:8080/metrics"]
  bearer_token = '/path/to/bearer/token'
  insecure_skip_verify = true
```

```toml
# Authorize using x509 certs
[[inputs.prometheus]]
  # An array of urls to scrape metrics from.
  urls = ["https://my-kube-apiserver:8080/metrics"]

  ssl_ca = '/path/to/cafile'
  ssl_cert = '/path/to/certfile'
  ssl_key = '/path/to/keyfile'
```

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

### Measurements & Fields & Tags:

Measurements and fields could be any thing.
It just depends of what you're quering.

Example:

```
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.00010425500000000001
go_gc_duration_seconds{quantile="0.25"} 0.000139108
go_gc_duration_seconds{quantile="0.5"} 0.00015749400000000002
go_gc_duration_seconds{quantile="0.75"} 0.000331463
go_gc_duration_seconds{quantile="1"} 0.000667154
go_gc_duration_seconds_sum 0.0018183950000000002
go_gc_duration_seconds_count 7
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 15
```

- go_goroutines,
    - gauge (integer, unit)
- go_gc_duration_seconds
    - field3 (integer, bytes)

- All measurements have the following tags:
    - url=http://my-kube-apiserver:8080/metrics
- go_goroutines has the following tags:
    - kubeservice=kube-apiserver
- go_gc_duration_seconds has the following tags:
    - kubeservice=kube-apiserver

### Example Output:

Example of output with configuration given above:

```
$ ./telegraf --config telegraf.conf --test
k8s_go_goroutines,kubeservice=kube-apiserver,url=http://my-kube-apiserver:8080/metrics gauge=536 1456857329391929813
k8s_go_gc_duration_seconds,kubeservice=kube-apiserver,url=http://my-kube-apiserver:8080/metrics 0=0.038002142,0.25=0.041732467,0.5=0.04336492,0.75=0.047271799,1=0.058295811,count=0,sum=208.334617406 1456857329391929813
```
