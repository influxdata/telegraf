# ServiceNow Metrics serializer

The ServiceNow Metrics serializer outputs metrics in the
[ServiceNow Operational Intelligence format][ServiceNow-format] or optionally
with the [ServiceNow JSONv2 format][ServiceNow-jsonv2]

It can be used to write to a file using the file output, or for sending metrics
to a MID Server with Enable REST endpoint activated using the standard telegraf
HTTP output. If you are using the HTTP output, this serializer knows how to
batch the metrics so you do not end up with an HTTP POST per metric.

[ServiceNow-format]: https://docs.servicenow.com/bundle/london-it-operations-management/page/product/event-management/reference/mid-POST-metrics.html
[ServiceNow-jsonv2]: https://docs.servicenow.com/bundle/tokyo-application-development/page/integrate/inbound-other-web-services/concept/c_JSONv2WebService.html

An example Operational Intelligence format event looks like:

```json
[
  {
    "metric_type": "Disk C: % Free Space",
    "resource": "C:\\",
    "node": "lnux100",
    "value": 50,
    "timestamp": 1473183012000,
    "ci2metric_id": {
        "node": "lnux100"
    },
    "source": "Telegraf"
  }
]
```

An example of the JSONv2 format even looks like:

```json
{
  "records": [
    {
      "metric_type": "Disk C: % Free Space",
      "resource": "C:\\",
      "node": "lnux100",
      "value": 50,
      "timestamp": 1473183012000,
      "ci2metric_id": {
        "node": "lnux100"
      },
      "source": "Telegraf"
    }
  ]
}
```

## Using with the HTTP output

To send this data to a ServiceNow MID Server with Web Server extension activated, you can use the HTTP output, there are some custom headers that you need to add to manage the MID Web Server authorization, here's a sample config for an HTTP output:

```toml
[[outputs.http]]
  ## URL is the address to send metrics to
  url = "http://<mid server fqdn or ip address>:9082/api/mid/sa/metrics"

  ## Timeout for HTTP message
  # timeout = "5s"

  ## HTTP method, one of: "POST" or "PUT"
  method = "POST"

  ## HTTP Basic Auth credentials
  username = 'evt.integration'
  password = 'P@$$w0rd!'

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
  data_format = "nowmetric"

  ## Format Type
  ## By default, the serializer returns an array of metrics matching the
  ## Now Metric Operational Intelligence format or with the option set to 'oi'.
  ## Optionally, if set to 'jsonv2' the output format will involve the newer
  ## JSON object based format.
  # nowmetric_format = "oi"

  ## Additional HTTP headers
  [outputs.http.headers]
  #   # Should be set manually to "application/json" for json data_format
  Content-Type = "application/json"
  Accept = "application/json"
```

Starting with the [London release](https://docs.servicenow.com/bundle/london-it-operations-management/page/product/event-management/task/event-rule-bind-metrics-to-host.html
),
you also need to explicitly create event rule to allow binding of metric events to host CIs.

## Metric Format

The following describes the two options of the `nowmetric_format` option:

The Operational Intelligence format is used along with the
`/api/mid/sa/metrics` API endpoint. The payload is requires a JSON array full
of metrics. This is the default settings and used when set to `oi`. See the
[ServiceNow KB0853084][KB0853084] for more details on this format.

Another option is the use of the [JSONv2 web service][jsonv2]. This service
requires a different format that is [JSON object based][jsonv2_format]. This
option is used when set to `jsonv2`.

[KB0853084]: https://support.servicenow.com/kb?id=kb_article_view&sysparm_article=KB0853084
[jsonv2]: https://docs.servicenow.com/bundle/tokyo-application-development/page/integrate/inbound-other-web-services/concept/c_JSONv2WebService.html
[jsonv2_format]: https://docs.servicenow.com/bundle/tokyo-application-development/page/integrate/inbound-other-web-services/concept/c_JSONObjectFormat.html

## Using with the File output

You can use the file output to output the payload in a file.
In this case, just add the following section to your telegraf config file

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["C:/Telegraf/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "nowmetric"
```
