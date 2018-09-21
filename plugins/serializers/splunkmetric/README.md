# Splunk Metrics serializer

The Splunk Metrics serializer outputs metrics in the [Splunk metric HEC JSON format][splunk-format].

It can be used to write to a file using the file output, or for sending metrics to a HEC using the standard telegraf HTTP output.
If you're using the HTTP output, this serializer knows how to batch the metrics so you don't end up with an HTTP POST per metric.

[splunk-format]: http://dev.splunk.com/view/event-collector/SP-CAAAFDN#json

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
   ## URL is the address to send metrics to
   url = "https://localhost:8088/services/collector"

   ## Timeout for HTTP message
   # timeout = "5s"

   ## HTTP method, one of: "POST" or "PUT"
   # method = "POST"

   ## HTTP Basic Auth credentials
   # username = "username"
   # password = "pa$$word"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false

   ## Data format to output.
   ## Each data format has it's own unique set of configuration options, read
   ## more about them here:
   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
   data_format = "splunkmetric"
    ## Provides time, index, source overrides for the HEC
   splunkmetric_hec_routing = true

   ## Additional HTTP headers
    [outputs.http.headers]
   # Should be set manually to "application/json" for json data_format
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

## Using with the File output

You can use the file output when running telegraf on a machine with a Splunk forwarder.

A sample event when `hec_routing` is false (or unset) looks like:
```javascript
{
    "_value": 0.6,
    "cpu": "cpu0",
    "dc": "mobile",
    "metric_name": "cpu.usage_user",
    "user": "ronnocol",
    "time": 1529708430
}
```
Data formatted in this manner can be ingested with a simple `props.conf` file that
looks like this:

```ini
[telegraf]
category = Metrics
description = Telegraf Metrics
pulldown_type = 1
DATETIME_CONFIG =
NO_BINARY_CHECK = true
SHOULD_LINEMERGE = true
disabled = false
INDEXED_EXTRACTIONS = json
KV_MODE = none
TIMESTAMP_FIELDS = time
TIME_FORMAT = %s.%3N
```

An example configuration of a file based output is:

```toml
 # Send telegraf metrics to file(s)
[[outputs.file]]
   ## Files to write to, "stdout" is a specially handled file.
   files = ["/tmp/metrics.out"]

   ## Data format to output.
   ## Each data format has its own unique set of configuration options, read
   ## more about them here:
   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
   data_format = "splunkmetric"
   hec_routing = false
```
