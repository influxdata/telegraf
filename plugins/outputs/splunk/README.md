# Splunk Output Plugin

This plugin writes metrics to a [Splunk Metric Index](https://docs.splunk.com/Documentation/Splunk/latest/Indexer/Setupmultipleindexes#Create_metrics_indexes) via a [Splunk HTTP Event Collector (HEC)](http://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector).  This requires an `Authorization Token` which will be created when you setup your Splunk HEC Token.

Notes:
1. If any point value cannot be converted to a float64, that metric will be skipped.
2. Metrics are grouped by converting any `_` characters to `.` in the Metric Name.



### Configuration:

```toml
# Configuration for Splunk server to send metrics to
[[outputs.splunk]]
  ## REQUIRED
  ## URL of the Splunk Enterprise HEC endpoint (i.e.: http://localhost:8088/services/collector)
  SplunkUrl = "http://localhost:8088/services/collector"

  ## REQUIRED
  ## Splunk Authorization Token for sending data to a Splunk HTTPEventCollector (HEC).
  ##   Note:  This Token should map to a 'metrics' index in Splunk.  
  AuthString = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"

  ## OPTIONAL:  prefix for metrics keys
  #prefix = "my.specific.prefix."

  ## OPTIONAL:  whether to use "value" for name of simple fields
  ##  Default is false which will result in using only the measurement name as the metric name, not "value"
  #simple_fields = false

  ## OPTIONAL:  character to use between metric and field name.  defaults to . (dot)
  #metric_separator = "."

  ## OPTIONAL:  Convert metric name paths to use metricSeperator character
  ## When true will convert all _ (underscore) chartacters in final metric name
  #convert_paths = false

  ## OPTIONAL:  Replace special characters in metric names with "-".  
  ## This can be useful if metric names contain special characters  
  #replace_special_chars = false

  ## OPTIONAL:  Use Regex to sanitize metric and tag names from invalid characters
  ## Regex is more thorough, but significantly slower
  #use_regex = false

  ## OPTIONAL:  whether to convert boolean values to numeric values, with false -> 0.0 and true -> 1.0.  default true
  #convert_bool = true
```


### Convert Path & Metric Separator
If the `convert_path` option is true any `_` in metric and field names will be converted to the `metric_separator` value. 
By default, the `convert_path` option is false and `metric_separator` is `.` (dot). 


### Replace Special Characters & Use Regex
Special characters in the metric name can be automatically converted to `-` by setting replace_special_chars to true.  This is false by default.  
The `use_regex` setting can be used to ensure all special characters are properly handled, but can lead to performance degradation.



### Splunk Metric Data format
The expected input for a Splunk Metric is:
```
{  "time":<timestamp>,
   "event":"metric",
   "source":"",
   "host":"<host>",
   "fields": {
      "metric_name":"<your.metric.name>,
      "_value":<your_metric_value>,
      "<dimension1_name>":"<dimension1_value>",
      "<dimension2_name>":"<dimension2_value>",
      ...,
      "<dimensionN_name>":"<dimensionN_value>"
   }
}
```
More information about the Splunk Metric data format is available [here](https://docs.splunk.com/Documentation/Splunk/latest/Metrics/Overview)

