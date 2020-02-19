# Azure Event Hubs input plugin

This plugin provides a consumer for use with Azure Event Hubs and Azure IoT Hub. The implementation is in essence a wrapper for [Microsoft Azure Event Hubs Client for Golang](https://github.com/Azure/azure-event-hubs-go).

## Configuration

```toml
[[inputs.eventhub]]
  ## The default behavior is to create a new Event Hub client from environment variables.
  ## This requires one of the following sets of environment variables to be set:
  ##
  ## 1) Expected Environment Variables:
  ##    - "EVENTHUB_NAMESPACE"
  ##    - "EVENTHUB_NAME"
  ##    - "EVENTHUB_CONNECTION_STRING"
  ##
  ## 2) Expected Environment Variables:
  ##    - "EVENTHUB_NAMESPACE"
  ##    - "EVENTHUB_NAME"
  ##    - "EVENTHUB_KEY_NAME"
  ##    - "EVENTHUB_KEY_VALUE"

  ## Uncommenting the option below will create an Event Hub client based solely on the connection string.
  ## This can either be the associated environment variable or hard coded directly.
  # connection_string = "$EVENTHUB_CONNECTION_STRING"

  ## Set persistence directory to a valid folder to use a file persister instead of an in-memory persister
  # persistence_dir = ""

  ## Change the default consumer group
  # consumer_group = ""

  ## By default the event hub receives all messages present on the broker.
  ## Alternative modes can be set below. The timestamp should be in RFC3339 format.
  ## The 3 options below only apply if no valid persister is read from memory or file (e.g. first run).
  # from_timestamp = ""
  # starting_offset = ""
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
  # max_undelivered_messages = 1000

  ## Prefix to use for the system properties of Event Hub and IoT Hub messages
  # system_properties_prefix = "SystemProperties."

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```
## Testing

The main focus for development of this plugin is Azure IoT hub:

1. Create an Azure IoT Hub by following any of the guides provided here: https://docs.microsoft.com/en-us/azure/iot-hub/
2. Create a device, for example a [simulated Raspberry Pi](https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-raspberry-pi-web-simulator-get-started)
3. The connection string needed for the plugin is located under *Shared access policies*, both the *iothubowner* and *service* policies should work