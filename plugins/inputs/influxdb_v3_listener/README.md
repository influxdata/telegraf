# InfluxDB V3 Listener Input Plugin

This plugin listens for requests sent according to the [InfluxDB 3 HTTP API][api]
and allows Telegraf to serve as a proxy or router for the `/api/v3/write_lp`
endpoint.

Line protocol received on that endpoint is parsed and passed on to the
configured outputs. Writes to the InfluxDB 1.x `/write` and InfluxDB 2.x
`/api/v2/write` endpoints are handled by the
[influxdb_listener][influxdb_listener] and
[influxdb_v2_listener][influxdb_v2_listener] plugins respectively.

⭐ Telegraf v1.40.0
🏷️ datastore
💻 all

[api]: https://docs.influxdata.com/influxdb3/core/api/
[influxdb_listener]: ../influxdb_listener/README.md
[influxdb_v2_listener]: ../influxdb_v2_listener/README.md

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listen and wait for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Tracking metric support <!-- @/docs/includes/plugin_tracking_metrics.md -->

This plugin supports [tracking metrics][METRICS.md], which allows the plugin
to be notified when metrics have been delivered to all outputs, enabling proper
acknowledgment back to the source.

[METRICS.md]: ../../../docs/METRICS.md#tracking-metrics

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret store support

This plugin supports secrets from secret stores for the `token` option.
See the [secret store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Accept metrics over the InfluxDB 3 HTTP API
[[inputs.influxdb_v3_listener]]
  ## Address and port to host the listener on
  service_address = ":8181"

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

  ## Optional tag to store the database of the write in.
  ## The default of an empty string will not record the database.
  # database_tag = ""

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

## Endpoints

| endpoint            | method     | description                        |
|---------------------|------------|------------------------------------|
| `/api/v3/write_lp`  | POST       | write line protocol                |
| `/health`           | GET        | health of the listener             |
| `/api/v1/health`    | GET        | health of the listener             |
| `/ping`             | GET, POST  | version information                |

If a `token` is configured, writes have to carry it in the `Authorization`
header as `Bearer <token>`, `Token <token>` or as the password part of a
`Basic` credential. The health and ping endpoints are always unauthenticated.

### Write parameters

The `/api/v3/write_lp` endpoint accepts the query parameters of the InfluxDB 3
write API:

| parameter        | default | description                                    |
|------------------|---------|------------------------------------------------|
| `db`             |         | database of the write, required                |
| `precision`      | `auto`  | timestamp precision of the line protocol       |
| `accept_partial` | `true`  | write the valid lines of a rejected batch      |
| `no_sync`        | `false` | do not wait for the write-ahead log            |

The `precision` parameter accepts `auto`, `second`, `millisecond`,
`microsecond` and `nanosecond` as well as the `s`, `ms`, `us`, `u`, `ns` and `n`
abbreviations. With `auto` the unit of each timestamp is determined by its
magnitude, the same way InfluxDB does it.

With `accept_partial` enabled, which is the default, the lines that do parse are
written and the request is answered with HTTP status 400 and the list of the
lines that did not

```json
{
  "error": "partial write of line protocol occurred",
  "data": [
    {
      "original_line": "cpu,host=b value=+Inf",
      "line_number": 2,
      "error_message": "metric parse error: expected field at 2:22"
    }
  ]
}
```

Disabling it rejects the whole batch on the first invalid line, writing nothing
at all and reporting only that line

```json
{
  "error": "line protocol parsing error",
  "data": {
    "original_line": "cpu,host=b value=+Inf",
    "line_number": 2,
    "error_message": "metric parse error: expected field at 2:22"
  }
}
```

## Deviations from the InfluxDB 3 API

This plugin implements the write API of InfluxDB 3, not InfluxDB 3 itself.
There is no catalog, no schema and no storage behind it, so it differs in the
following points:

- The `no_sync` parameter is accepted and ignored, as there is no write-ahead
  log to acknowledge.
- Writes are only rejected for line protocol parse errors. InfluxDB also
  rejects lines on schema conflicts, column limits and retention periods, none
  of which exist here, so those `error_message` values never occur. The
  messages of the parse errors are the ones of the Telegraf parser and do not
  match the InfluxDB wording either.
- Database names are only required to be non-empty, the character and length
  restrictions of InfluxDB are not enforced.
- The query, catalog and processing engine endpoints are not served, a request
  to any of them is answered with HTTP status 404. Requesting a served endpoint
  with the wrong method gives HTTP status 405 rather than 404.
- Writes without a timestamp are stamped with the current time in nanosecond
  precision. InfluxDB truncates that time to the requested precision.
- The health endpoint reports HTTP status 503 once `max_undelivered_metrics`
  pending metrics have not been delivered to the outputs yet.

## Metrics

Metrics are created from InfluxDB Line Protocol in the request body.

## Example Output

Using

```sh
curl -i -XPOST 'http://localhost:8181/api/v3/write_lp?db=mydb' --data-binary 'cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000'
```

will produce the following metric

```text
cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000
```
