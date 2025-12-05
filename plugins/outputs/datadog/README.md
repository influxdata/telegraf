# Datadog Output Plugin

This plugin writes metrics to the [Datadog Metrics API][metrics] and requires an
`apikey` which can be obtained on the [website][apikey] for the account.

> [!NOTE]
> This plugin supports the v1 API.

⭐ Telegraf v0.1.6
🏷️ applications, cloud, datastore
💻 all

[metrics]: https://docs.datadoghq.com/api/v1/metrics/#submit-metrics
[apikey]: https://app.datadoghq.com/account/settings#api

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

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

  ## When non-zero, converts count metrics submitted by inputs.statsd
  ## into rate, while dividing the metric value by this number.
  ## Note that in order for metrics to be submitted simultaenously alongside
  ## a Datadog agent, rate_interval has to match the interval used by the
  ## agent - which defaults to 10s
  # rate_interval = 0s
```

## Metrics

Datadog metric names are formed by joining the Telegraf metric name and the
field key with a `.` character.

Field values are converted to floating point numbers.  Strings and floats that
cannot be sent over JSON, namely NaN and Inf, are ignored.

Setting `rate_interval` to non-zero will convert `count` metrics to `rate`
and divide its value by this interval before submitting to Datadog.
This allows Telegraf to submit metrics alongside Datadog agents when their rate
intervals are the same (Datadog defaults to `10s`).
Note that this only supports metrics ingested via `inputs.statsd` given
the dependency on the `metric_type` tag it creates. There is only support for
`counter` metrics, and `count` values from `timing` and `histogram` metrics.
