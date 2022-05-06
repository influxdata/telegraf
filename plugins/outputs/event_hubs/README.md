# Azure Event Hubs Output Plugin

This plugin for [Azure Event
Hubs](https://azure.microsoft.com/en-gb/services/event-hubs/) will send metrics
to a single Event Hub within an Event Hubs namespace. Metrics are sent as
message batches, each message payload containing one metric object. The messages
do not specify a partition key, and will thus be automatically load-balanced
(round-robin) across all the Event Hub partitions.

## Metrics

The plugin uses the Telegraf serializers to format the metric data sent in the
message payloads. You can select any of the supported output formats, although
JSON is probably the easiest to integrate with downstream components.

## Configuration

```toml
# Configuration for Event Hubs output plugin
[[outputs.event_hubs]]
  ## The full connection string to the Event Hub (required)
  ## The shared access key must have "Send" permissions on the target Event Hub.
  connection_string = "Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=superSecret1234=;EntityPath=hubName"
  ## Client timeout (defaults to 30s)
  # timeout = "30s"
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "json"
```
