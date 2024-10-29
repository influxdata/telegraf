# New Relic Output Plugin

This plugins writes metrics to [New Relic Insights][newrelic] using the
[Metrics API][metrics_api]. To use this plugin you have to obtain an
[Insights API Key][insights_api_key].

⭐ Telegraf v1.15.0
🏷️ applications
💻 all

[newrelic]: https://newrelic.com
[metrics_api]: https://docs.newrelic.com/docs/data-ingest-apis/get-data-new-relic/metric-api/introduction-metric-api
[insights_api_key]: https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys#user-api-key

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send metrics to New Relic metrics endpoint
[[outputs.newrelic]]
  ## The 'insights_key' parameter requires a NR license key.
  ## New Relic recommends you create one
  ## with a convenient name such as TELEGRAF_INSERT_KEY.
  ## reference: https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/#ingest-license-key
  # insights_key = "New Relic License Key Here"

  ## Prefix to add to add to metric name for easy identification.
  ## This is very useful if your metric names are ambiguous.
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
