# InfluxDB V2 Listener Input Plugin

This plugin listens for requests sent according to the
[InfluxDB HTTP v2 API][influxdb_http_api]. This allows Telegraf to serve as a
proxy/router for the `/api/v2/write` endpoint of the InfluxDB HTTP API.

The `/api/v2/write` endpoint supports the `precision` query parameter and can be
set to one of `ns`, `us`, `ms`, `s`.  All other parameters are ignored and defer
to the output plugins configuration.

⭐ Telegraf v1.16.0
🏷️ datastore
💻 all

[influxdb_http_api]: https://docs.influxdata.com/influxdb/v2/api/

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Accept metrics over InfluxDB 2.x HTTP API
[[inputs.influxdb_v2_listener]]
  ## Address and port to host InfluxDB listener on
  ## (Double check the port. Could be 9999 if using OSS Beta)
  service_address = ":8086"

  ## Maximum undelivered metrics before rate limit kicks in.
  ## When the rate limit kicks in, HTTP status 429 will be returned.
  ## 0 disables rate limiting
  # max_undelivered_metrics = 0

  ## Maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## Maximum duration before timing out write of the response
  # write_timeout = "10s"

  ## Maximum allowed HTTP request body size in bytes.
  ## 0 means to use the default of 32MiB.
  # max_body_size = "32MiB"

  ## Optional tag to determine the bucket.
  ## If the write has a bucket in the query string then it will be kept in this tag name.
  ## This tag can be used in downstream outputs.
  ## The default value of nothing means it will be off and the database will not be recorded.
  # bucket_tag = ""

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Optional token to accept for HTTP authentication.
  ## You probably want to make sure you have TLS configured above for this.
  # token = "some-long-shared-secret-token"

  ## Influx line protocol parser
  ## 'internal' is the default. 'upstream' is a newer parser that is faster
  ## and more memory efficient.
  # parser_type = "internal"
```

## Metrics

Metrics are created from InfluxDB Line Protocol in the request body.

## Example Output

Using

```sh
curl -i -XPOST 'http://localhost:8186/api/v2/write' --data-binary 'cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000'
```

will produce the following metric

```text
cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000
```
