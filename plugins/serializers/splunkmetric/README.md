# Splunk Metrics serialzier

This serializer formats and outputs the metric data in a format that can be consumed by a Splunk metrics index. It can be used to write to a file using the file output, or for sending metrics to a HEC using the standard telegraf HTTP output. If you're using the HTTP output, this serializer knows how to batch the metrics so you don't end up with an HTTP POST per metric.

Th data is output in a format that conforms to the specified Splunk HEC JSON format as found here: [Send metrics in JSON format](http://dev.splunk.com/view/event-collector/SP-CAAAFDN).

An example event looks like:
```javascript
{
  "time": 1529708430,
  "event": "metric",
  "host": "patas-mbp",
  "fields": {
    "_value": 0.6,
    "cpu": "cpu0",
    "dc": "mobile",
    "metric_name": "cpu.usage_user",
    "user": "ronnocol"
  }
}
```
In the above snippet, the following keys are dimensions:
* cpu
* dc
* user

## Using with the HTTP output

To send this data to a Splunk HEC, you can use the HTTP output, there are some custom headers that you need to add
to manage the HEC authorization, here's a sample config for an HTTP output:

```toml
[[outputs.http]]
#   ## URL is the address to send metrics to
   url = "https://localhost:8088/services/collector"
#
#   ## Timeout for HTTP message
#   # timeout = "5s"
#
#   ## HTTP method, one of: "POST" or "PUT"
#   # method = "POST"
#
#   ## HTTP Basic Auth credentials
#   # username = "username"
#   # password = "pa$$word"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
#
#   ## Data format to output.
#   ## Each data format has it's own unique set of configuration options, read
#   ## more about them here:
#   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
   data_format = "splunkmetric"
#
#   ## Additional HTTP headers
    [outputs.http.headers]
#   # Should be set manually to "application/json" for json data_format
      Content-Type = "application/json"
      Authorization = "Splunk xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      X-Splunk-Request-Channel = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
```

## Overrides
You can override the default values for the HEC token you are using by adding additional tags to the config file.

The following aspects of the token can be overriden with tags:
* index
* source

You can either use `[global_tags]` or using a more advanced configuration as documented [here](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md).
 
Such as this example which overrides the index just on the cpu metric:
```toml
[[inputs.cpu]]
  percpu = false
  totalcpu = true
  [inputs.cpu.tags]
    index = "cpu_metrics"
```

