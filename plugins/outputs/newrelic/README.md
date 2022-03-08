# New Relic output plugin

This plugins writes to New Relic Insights using the [Metrics API][].

To use this plugin you must first obtain an [Insights API Key][].

Telegraf minimum version: Telegraf 1.15.0

## Configuration

```toml
[[outputs.newrelic]]
  ## The 'insights_key' parameter requires a NR license key.
  ## New Relic recommends you create one
  ## with a convenient name such as TELEGRAF_INSERT_KEY.
  ## reference: https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/#ingest-license-key
  insights_key = "license key here"

  ## Prefix to add to add to metric name for easy identification.
  # metric_prefix = ""

  ## Timeout for writes to the New Relic API.
  # timeout = "15s"

  ## HTTP Proxy override. If unset use values from the standard
  ## proxy environment variables to determine proxy, if any.
  # http_proxy = "http://corporate.proxy:3128"

  ## Metric URL override to enable geographic location endpoints.
  # If not set use values from the standard
  # metric_url = "https://metric-api.newrelic.com/metric/v1"
```

[Metrics API]: https://docs.newrelic.com/docs/data-ingest-apis/get-data-new-relic/metric-api/introduction-metric-api
[Insights API Key]: https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys#user-api-key
