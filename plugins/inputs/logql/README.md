# LogQL Input Plugin

This plugin gathers metrics from a [Loki][loki] endpoint using
[LogQL queries][logql] via the [HTTP API][http_api].

⭐ Telegraf v1.37.0
🏷️ datastore
💻 all

[loki]: https://grafana.com/oss/loki/
[logql]: https://grafana.com/docs/loki/latest/query/
[http_api]: https://grafana.com/docs/loki/latest/reference/loki-http-api/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username`, `password`
and `token` option. See the [secret-store documentation][SECRETSTORE] for
more details on how to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Query Loki endpoints using LogQL
[[inputs.logql]]
  ## URL of the Loki endpoint
  # url = "http://localhost:3100"

  ## Basic authentication properties
  # username = ""
  # password = ""

  ## Bearer token based authentication
  # token = ""

  ## Organization IDs for the queries in multi-tenant setups
  # organizations = []

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
  # [[inputs.logql.instant]]
  #   ## Fallback name of the resulting metrics to use as metric name in case
  #   ## the __name__ property of the query results is empty.
  #   # name = "logql"
  #
  #   ## Query to execute
  #   query = 'count_over_time({job="varlogs"}[24h])'
  #
  #   ## Sorting direction of logs, either "forward" or "backward"
  #   # sorting = "backward"
  #
  #   ## Limit for the number of results returned by the server with zero
  #   ## meaning no limit
  #   # limit = 0

  ## Range queries, multiple instances are allowed
  # [[inputs.logql.range]]
  #   ## Fallback name of the resulting metrics to use as metric name in case
  #   ## the __name__ property of the query results is empty.
  #   # name = "logql"
  #
  #   ## Query to execute
  #   query = '{job="varlogs"}'
  #
  #   ## Range parameters relative to the gathering time with positive values
  #   ## refer to BEFORE and negative to AFTER the gathering time
  #   start = "5m"
  #   # end = "0s"
  #   # step = "0s"
  #   # interval = "0s"
  #
  #   ## Sorting direction of logs, either "forward" or "backward"
  #   # sorting = "backward"
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

### Vector and Matrix Results

Vector and metric results will produce one or more metrics with the metric named
being either `logql` or the `name` specified in the query.
Labels of the results will be kept as tags if they are not internal i.e. they
are not starting and ending with a double underscore (`__<label>__`).
The returned value for each sample will be stored in a field called `value` and
the provided Loki timestamp will be used as metric timestamp.

### Stream Results

A stream result will produce one or more metrics with the metric named being
either `logql` or the `name` specified in the query.
Labels of the results will be kept as tags if they are not internal i.e. they
are not starting and ending with a double underscore (`__<label>__`).
The returned log-line for each sample will be stored in a field called `message`
and the provided Loki timestamp will be used as metric timestamp.

## Example Output

For example, the instant query

```logql
sum(rate({job="varlogs"}[5m])) by (detected_level)
```

returns

```text
logql,detected_level=error,host=Hugin value=0.5833333333333334 1762943358000000000
logql,detected_level=info,host=Hugin value=45.8 1762943358000000000
logql,detected_level=unknown,host=Hugin value=741.71 1762943358000000000
logql,detected_level=warn,host=Hugin value=16.566666666666666 1762943358000000000
```

The example range-query

```logql
count by(detected_level) (rate({job="varlogs", filename="/var/log/Xorg.0.log"} [5m]))
```

with a range of 30 minutes ago to now and one minute stepping results in

```text
logql,detected_level=error,host=Hugin value=1 1762943220000000000
logql,detected_level=error,host=Hugin value=1 1762943280000000000
logql,detected_level=error,host=Hugin value=1 1762943340000000000
logql,detected_level=error,host=Hugin value=1 1762943400000000000
logql,detected_level=error,host=Hugin value=1 1762943460000000000
logql,detected_level=info,host=Hugin value=1 1762943220000000000
logql,detected_level=info,host=Hugin value=1 1762943280000000000
logql,detected_level=info,host=Hugin value=1 1762943340000000000
logql,detected_level=info,host=Hugin value=1 1762943400000000000
logql,detected_level=info,host=Hugin value=1 1762943460000000000
logql,detected_level=unknown,host=Hugin value=1 1762943220000000000
logql,detected_level=unknown,host=Hugin value=1 1762943280000000000
logql,detected_level=unknown,host=Hugin value=1 1762943340000000000
logql,detected_level=unknown,host=Hugin value=1 1762943400000000000
logql,detected_level=unknown,host=Hugin value=1 1762943460000000000
```

When quering raw log entries (i.e. stream results) e.g. via

```logql
{job="varlogs", filename="/var/log/Xorg.0.log"}
```

with sorting forward and limiting the number of returned values to 5 you get

```text
logql,detected_level=unknown,filename=/var/log/Xorg.0.log,host=Hugin,job=varlogs,service_name=varlogs message="[     6.806] (--) Log file renamed from \"/var/log/Xorg.pid-693.log\" to \"/var/log/Xorg.0.log\"" 1762943173000000000
logql,detected_level=unknown,filename=/var/log/Xorg.0.log,host=Hugin,job=varlogs,service_name=varlogs message="[     6.807] " 1762943173000000000
logql,detected_level=unknown,filename=/var/log/Xorg.0.log,host=Hugin,job=varlogs,service_name=varlogs message="X.Org X Server 1.21.1.14" 1762943173000000000
logql,detected_level=unknown,filename=/var/log/Xorg.0.log,host=Hugin,job=varlogs,service_name=varlogs message="X Protocol Version 11, Revision 0" 1762943173000000000
logql,detected_level=unknown,filename=/var/log/Xorg.0.log,host=Hugin,job=varlogs,service_name=varlogs message="[     6.807] Current Operating System: Linux Hugin 6.12.1-arch1-1 #1 SMP PREEMPT_DYNAMIC Fri, 22 Nov 2024 16:04:27 +0000 x86_64" 1762943173000000000
```
