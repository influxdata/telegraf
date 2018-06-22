# Splunk HEC Output Plugin

This plugin writes to the [Splunk HEC API](http://dev.splunk.com/view/event-collector/SP-CAAAFDN)
and requires an `token` which can be obtained by following instructions [here](https://docs.splunk.com/Documentation/Splunk/7.0.1/Metrics/GetMetricsInOther#Get_metrics_in_from_clients_over_HTTP_or_HTTPS)
for your Splunk Enterprise installation.

If the value being sent cannot be converted to a float64, the metric is skipped.
