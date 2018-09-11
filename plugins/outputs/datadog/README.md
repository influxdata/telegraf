# Datadog Output Plugin

This plugin writes to the [Datadog Metrics API](http://docs.datadoghq.com/api/#metrics)
and requires an `apikey` which can be obtained [here](https://app.datadoghq.com/account/settings#api)
for the account.
Application keys, in conjunction with your API key are required to write metrics through
Datadog's programmatic API.

If the point value being sent cannot be converted to a float64, the metric is skipped.

Metrics are grouped by converting any `_` characters to `.` in the Point Name.