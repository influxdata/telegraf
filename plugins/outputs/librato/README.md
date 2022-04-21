# Librato Output Plugin

This plugin writes to the [Librato Metrics API][metrics-api] and requires an
`api_user` and `api_token` which can be obtained [here][tokens] for the account.

The `source_tag` option in the Configuration file is used to send contextual
information from Point Tags to the API.

If the point value being sent cannot be converted to a float64, the metric is
skipped.

Currently, the plugin does not send any associated Point Tags.

[metrics-api]: http://dev.librato.com/v1/metrics#metrics

[tokens]: https://metrics.librato.com/account/api_tokens

## Configuration

```toml
# Configuration for Librato API to send metrics to.
[[outputs.librato]]
  ## Librato API Docs
  ## http://dev.librato.com/v1/metrics-authentication
  ## Librato API user
  api_user = "telegraf@influxdb.com" # required.
  ## Librato API token
  api_token = "my-secret-token" # required.
  ## Debug
  # debug = false
  ## Connection timeout.
  # timeout = "5s"
  ## Output source Template (same as graphite buckets)
  ## see https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#graphite
  ## This template is used in librato's source (not metric's name)
  template = "host"
```
