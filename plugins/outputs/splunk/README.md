# Splunk Output Plugin

This plugin writes to a [Splunk HTTP Event Collector (HEC)](http://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector)
and requires an `Authorization Token` which will be created when you setup your Splunk HEC Token.


If the point value being sent cannot be converted to a float64, the metric is skipped.

Metrics are grouped by converting any `_` characters to `.` in the Metric Name.