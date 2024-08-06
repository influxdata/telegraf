# Datadog Output Plugin

This plugin writes to the [Datadog Metrics API][metrics] and requires an
`apikey` which can be obtained [here][apikey] for the account. This plugin
supports the v1 API.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Configuration for DataDog API to send metrics to.
[[outputs.datadog]]
  ## Datadog API key
  apikey = "my-secret-key"

  ## Connection timeout.
  # timeout = "5s"

  ## Write URL override; useful for debugging.
  ## This plugin only supports the v1 API currently due to the authentication
  ## method used.
  # url = "https://app.datadoghq.com/api/v1/series"

  ## Set http_proxy
  # use_system_proxy = false
  # http_proxy_url = "http://localhost:8888"

  ## Override the default (none) compression used to send data.
  ## Supports: "zlib", "none"
  # compression = "none"

  ## Convert counts to rates
  ## Use this to be able to submit metrics from telegraf alongside Datadog agent
  # should_rate_counts = false

  ## Overrides the default rate interval used to divide count metrics by
  ## when should_rate_counts is enabled
  # rate_interval = 10
```

## Metrics

Datadog metric names are formed by joining the Telegraf metric name and the
field key with a `.` character.

Field values are converted to floating point numbers.  Strings and floats that
cannot be sent over JSON, namely NaN and Inf, are ignored.

Enabling the `should_rate_counts` will convert `count` metrics to `rate`
and divide it by the `rate_interval` before submitting to Datadog.
This allows telegraf to submit metrics alongside Datadog agents.
Note that this only supports metrics ingested via `inputs.statsd` given
the dependency on the `metric_type` tag it creates. There is only support for
`counter` metrics, and `count` values from `timing` and `histogram` metrics.

[metrics]: https://docs.datadoghq.com/api/v1/metrics/#submit-metrics
[apikey]: https://app.datadoghq.com/account/settings#api
