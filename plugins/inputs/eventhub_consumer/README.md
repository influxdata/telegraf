# Event Hub Consumer Input Plugin

This plugin provides a consumer for use with Azure Event Hubs and Azure IoT Hub.

## IoT Hub Setup

The main focus for development of this plugin is Azure IoT hub:

1. Create an Azure IoT Hub by following any of the guides provided here: [Azure
   IoT Hub](https://docs.microsoft.com/en-us/azure/iot-hub/)
2. Create a device, for example a [simulated Raspberry
   Pi](https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-raspberry-pi-web-simulator-get-started)
3. The connection string needed for the plugin is located under *Shared access
   policies*, both the *iothubowner* and *service* policies should work

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Azure Event Hubs service input plugin
[[inputs.eventhub_consumer]]
  ## The default behavior is to create a new Event Hub client from environment variables.
  ## This requires one of the following sets of environment variables to be set:
  ##
  ## 1) Expected Environment Variables:
  ##    - "EVENTHUB_CONNECTION_STRING"
  ##
  ## 2) Expected Environment Variables:
  ##    - "EVENTHUB_NAMESPACE"
  ##    - "EVENTHUB_NAME"
  ##    - "EVENTHUB_KEY_NAME"
  ##    - "EVENTHUB_KEY_VALUE"

  ## 3) Expected Environment Variables:
  ##    - "EVENTHUB_NAMESPACE"
  ##    - "EVENTHUB_NAME"
  ##    - "AZURE_TENANT_ID"
  ##    - "AZURE_CLIENT_ID"
  ##    - "AZURE_CLIENT_SECRET"

  ## Uncommenting the option below will create an Event Hub client based solely on the connection string.
  ## This can either be the associated environment variable or hard coded directly.
  ## If this option is uncommented, environment variables will be ignored.
  ## Connection string should contain EventHubName (EntityPath)
  # connection_string = ""

  ## Set persistence directory to a valid folder to use a file persister instead of an in-memory persister
  # persistence_dir = ""

  ## Change the default consumer group
  # consumer_group = ""

  ## By default the event hub receives all messages present on the broker, alternative modes can be set below.
  ## The timestamp should be in https://github.com/toml-lang/toml#offset-date-time format (RFC 3339).
  ## The 3 options below only apply if no valid persister is read from memory or file (e.g. first run).
  # from_timestamp =
  # latest = true

  ## Set a custom prefetch count for the receiver(s)
  # prefetch_count = 1000

  ## Add an epoch to the receiver(s)
  # epoch = 0

  ## Change to set a custom user agent, "telegraf" is used by default
  # user_agent = "telegraf"

  ## To consume from a specific partition, set the partition_ids option.
  ## An empty array will result in receiving from all partitions.
  # partition_ids = ["0","1"]

  ## Max undelivered messages
  ## This plugin uses tracking metrics, which ensure messages are read to
  ## outputs before acknowledging them to the original broker to ensure data
  ## is not lost. This option sets the maximum messages to read from the
  ## broker that have not been written by an output.
  ##
  ## This value needs to be picked with awareness of the agent's
  ## metric_batch_size value as well. Setting max undelivered messages too high
  ## can result in a constant stream of data batches to the output. While
  ## setting it too low may never flush the broker's messages.
  # max_undelivered_messages = 1000

  ## Set either option below to true to use a system property as timestamp.
  ## You have the choice between EnqueuedTime and IoTHubEnqueuedTime.
  ## It is recommended to use this setting when the data itself has no timestamp.
  # enqueued_time_as_ts = true
  # iot_hub_enqueued_time_as_ts = true

  ## Tags or fields to create from keys present in the application property bag.
  ## These could for example be set by message enrichments in Azure IoT Hub.
  # application_property_tags = []
  # application_property_fields = []

  ## Tag or field name to use for metadata
  ## By default all metadata is disabled
  # sequence_number_field = "SequenceNumber"
  # enqueued_time_field = "EnqueuedTime"
  # offset_field = "Offset"
  # partition_id_tag = "PartitionID"
  # partition_key_tag = "PartitionKey"
  # iot_hub_device_connection_id_tag = "IoTHubDeviceConnectionID"
  # iot_hub_auth_generation_id_tag = "IoTHubAuthGenerationID"
  # iot_hub_connection_auth_method_tag = "IoTHubConnectionAuthMethod"
  # iot_hub_connection_module_id_tag = "IoTHubConnectionModuleID"
  # iot_hub_enqueued_time_field = "IoTHubEnqueuedTime"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Environment Variables

[Full documentation of the available environment variables][envvar].

[envvar]: https://github.com/Azure/azure-event-hubs-go#environment-variables

## Metrics

## Example Output
