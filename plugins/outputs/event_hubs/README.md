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

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Configuration for Event Hubs output plugin
[[outputs.event_hubs]]
  ## The full connection string to the Event Hub (required)
  ## The shared access key must have "Send" permissions on the target Event Hub.
  connection_string = "Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=superSecret1234=;EntityPath=hubName"

  ## Client timeout (defaults to 30s)
  # timeout = "30s"

  ## Partition key
  ## Metric tag or field name to use for the event partition key. The value of
  ## this tag or field is set as the key for events if it exists. If both, tag
  ## and field, exist the tag is preferred.
  # partition_key = ""

  ## Set the maximum batch message size in bytes
  ## The allowable size depends on the Event Hub tier
  ## See: https://learn.microsoft.com/azure/event-hubs/event-hubs-quotas#basic-vs-standard-vs-premium-vs-dedicated-tiers
  ## Setting this to 0 means using the default size from the Azure Event Hubs Client library (1000000 bytes)
  # max_message_size = 1000000

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "json"
```
