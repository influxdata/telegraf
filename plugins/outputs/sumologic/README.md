# Sumo Logic Output Plugin

This plugin sends metrics to [Sumo Logic HTTP Source][http-source] in HTTP
messages, encoded using one of the output data formats.

Telegraf minimum version: Telegraf 1.16.0

Currently metrics can be sent using one of the following data formats, supported
by Sumologic HTTP Source:

* `graphite` - for Content-Type of `application/vnd.sumologic.graphite`
* `carbon2` - for Content-Type of `application/vnd.sumologic.carbon2`
* `prometheus` - for Content-Type of `application/vnd.sumologic.prometheus`

[http-source]: https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Metrics-to-an-HTTP-Source

## Configuration

```toml
# A plugin that can send metrics to Sumo Logic HTTP metric collector.
[[outputs.sumologic]]
  ## Unique URL generated for your HTTP Metrics Source.
  ## This is the address to send metrics to.
  # url = "https://events.sumologic.net/receiver/v1/http/<UniqueHTTPCollectorCode>"

  ## Data format to be used for sending metrics.
  ## This will set the "Content-Type" header accordingly.
  ## Currently supported formats:
  ## * graphite - for Content-Type of application/vnd.sumologic.graphite
  ## * carbon2 - for Content-Type of application/vnd.sumologic.carbon2
  ## * prometheus - for Content-Type of application/vnd.sumologic.prometheus
  ##
  ## More information can be found at:
  ## https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Metrics-to-an-HTTP-Source#content-type-headers-for-metrics
  ##
  ## NOTE:
  ## When unset, telegraf will by default use the influx serializer which is currently unsupported
  ## in HTTP Source.
  data_format = "carbon2"

  ## Timeout used for HTTP request
  # timeout = "5s"

  ## Max HTTP request body size in bytes before compression (if applied).
  ## By default 1MB is recommended.
  ## NOTE:
  ## Bear in mind that in some serializer a metric even though serialized to multiple
  ## lines cannot be split any further so setting this very low might not work
  ## as expected.
  # max_request_body_size = 1000000

  ## Additional, Sumo specific options.
  ## Full list can be found here:
  ## https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Metrics-to-an-HTTP-Source#supported-http-headers

  ## Desired source name.
  ## Useful if you want to override the source name configured for the source.
  # source_name = ""

  ## Desired host name.
  ## Useful if you want to override the source host configured for the source.
  # source_host = ""

  ## Desired source category.
  ## Useful if you want to override the source category configured for the source.
  # source_category = ""

  ## Comma-separated key=value list of dimensions to apply to every metric.
  ## Custom dimensions will allow you to query your metrics at a more granular level.
  # dimensions = ""
```
