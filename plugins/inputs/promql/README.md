# PromQL Input Plugin

This plugin gathers metrics from a [Prometheus][prometheus] endpoint using
[PromQL queries][promql] via the [HTTP API][http_api].

⭐ Telegraf v1.37.0
🏷️ datastore
💻 all

[prometheus]: https://prometheus.io/
[promql]: https://prometheus.io/docs/prometheus/latest/querying/basics/
[http_api]: https://prometheus.io/docs/prometheus/latest/querying/api/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username`, `password`
and `token` option. See the [secret-store documentation][SECRETSTORE] for
more details on how to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Query prometheus endpoints using PromQL
[[inputs.promql]]
  ## URL of the prometheus endpoint
  url = "http://localhost:9090"

  ## Basic authentication properties
  # username = ""
  # password = ""

  ## Bearer token based authentication
  # token = ""

  ## Timeout for executing queries with zero meaning no timeout
  # timeout = "5s"

  ## HTTP connection settings
  # idle_conn_timeout = "0s"
  # max_idle_conn = 0
  # max_idle_conn_per_host = 0
  # response_timeout = "0s"

  ## Optional proxy settings
  # use_system_proxy = false
  # http_proxy_url = ""

  ## Optional TLS settings
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable =
  ## Trusted root certificates for server
  # tls_ca = "/path/to/cafile"
  ## Used for TLS client certificate authentication
  # tls_cert = "/path/to/certfile"
  ## Used for TLS client certificate authentication
  # tls_key = "/path/to/keyfile"
  ## Password for the key file if it is encrypted
  # tls_key_pwd = ""
  ## Send the specified TLS server name via SNI
  # tls_server_name = "kubernetes.example.com"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## List of ciphers to accept, by default all secure ciphers will be accepted
  ## See https://pkg.go.dev/crypto/tls#pkg-constants for supported values.
  ## Use "all", "secure" and "insecure" to add all support ciphers, secure
  ## suites or insecure suites respectively.
  # tls_cipher_suites = ["secure"]
  ## Renegotiation method, "never", "once" or "freely"
  # tls_renegotiation_method = "never"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Instant queries, multiple instances are allowed
  # [[inputs.promql.instant]]
  #   ## Fallback name of the resulting metrics to use as metric name in case
  #   ## the __name__ property of the query results is empty.
  #   # name = "promql"
  #
  #   ## Query to execute
  #   query = 'prometheus_http_requests_total'
  #
  #   ## Limit for the number of results returned by the server with zero
  #   ## meaning no limit
  #   # limit = 0

  ## Range queries, multiple instances are allowed
  # [[inputs.promql.range]]
  #   ## Fallback name of the resulting metrics to use as metric name in case
  #   ## the __name__ property of the query results is empty.
  #   # name = "promql"
  #
  #   ## Query to execute
  #   query = 'prometheus_http_requests_total{job="prometheus"}'
  #
  #   ## Range parameters relative to the gathering time with positive values
  #   ## refer to BEFORE and negative to AFTER the gathering time
  #   start = "5m"
  #   # end = "0s"
  #   step = "1m"
  #
  #   ## Limit for the number of results returned by the server with zero
  #   ## meaning no limit
  #   # limit = 0
```

> [!NOTE]
> You can either use no authentication _or_ basic authentication _or_ Bearer
> token based authentication. Uncommenting both basic and Bearer token based
> authentication will fail.

## Metrics

The metrics collected by this input plugin will depend on the specified queries.
However, the resulting metrics will have the following structure for the
returned results.

### Scalar and String Results

A scalar or string result will produce single metrics named after the value of
the Prometheus `__name__` label. Other labels will be kept as tags. The
resulting metric uses the Prometheus timestamp and will have the value stored in
a `value` field.

### Vector and Matrix Results

A vector result will produce one or more metrics with the metric named after the
value of the Prometheus `__name__` label for each element of the vector. Other
labels will be kept as tags. All metrics will use the Prometheus timestamp.

Non-histogram results will have the value stored in a `value` field. Histogram
results will contain multiple fields with the field name being the upper bound
of the bin and a value with the bin count. Additionally, the metric will have a
`count` and a `sum` field.

## Example Output

For example, a range-query for
`prometheus_http_requests_total{job="prometheus", handler="/api/v1/query"}`
starting 5 minutes in the past with 1 minute stepping returns

```text
prometheus_http_requests_total,app=prometheus,code=200,handler=/api/v1/query,instance=localhost:9090,job=prometheus value=28 1758806201000000000
prometheus_http_requests_total,app=prometheus,code=200,handler=/api/v1/query,instance=localhost:9090,job=prometheus value=28 1758806261000000000
prometheus_http_requests_total,app=prometheus,code=200,handler=/api/v1/query,instance=localhost:9090,job=prometheus value=28 1758806321000000000
prometheus_http_requests_total,app=prometheus,code=200,handler=/api/v1/query,instance=localhost:9090,job=prometheus value=28 1758806381000000000
prometheus_http_requests_total,app=prometheus,code=200,handler=/api/v1/query,instance=localhost:9090,job=prometheus value=28 1758806441000000000
prometheus_http_requests_total,app=prometheus,code=200,handler=/api/v1/query,instance=localhost:9090,job=prometheus value=28 1758806501000000000
```
