# Librato Output Plugin

This plugin writes to the [Librato Metrics API](http://dev.librato.com/v1/metrics#metrics)
and requires an `api_user` and `api_token` which can be obtained [here](https://metrics.librato.com/account/api_tokens)
for the account.

The `source_tag` option in the Configuration file is used to send contextual information from
Point Tags to the API.

If the point value being sent cannot be converted to a float64, the metric is skipped.

Currently, the plugin does not send any associated Point Tags.