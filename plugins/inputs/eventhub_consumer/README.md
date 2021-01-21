# Event Hub Consumer Input Plugin

This plugin provides a consumer for use with Azure Event Hubs and Azure IoT Hub.

### IoT Hub Setup

The main focus for development of this plugin is Azure IoT hub:

1. Create an Azure IoT Hub by following any of the guides provided here: https://docs.microsoft.com/en-us/azure/iot-hub/
2. Create a device, for example a [simulated Raspberry Pi](https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-raspberry-pi-web-simulator-get-started)
3. The connection string needed for the plugin is located under _Shared access policies_, both the _iothubowner_ and _service_ policies should work

### Configuration

```toml
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
  data_format = "json"
```

#### Environment Variables

[Full documentation of the available environment variables][envvar].

[envvar]: https://github.com/Azure/azure-event-hubs-go#environment-variables

### Example

If we stream diagnostic setting 'AllMetrics' from one of Azure services to Eventhub,
our EventHub data may look like this:

```json
{
  "records": [
    {
      "count": 3,
      "total": 6,
      "average": 2,
      "resourceId": "/SUBSCRIPTIONS/123-**-456/LONG-STRING-HERE",
      "time": "2020-12-03T06:46:00.0000000Z",
      "metricName": "someMetricNameHere",
      "timeGrain": "PT1M"
    }
  ]
}
```

If so, our plugin configuration (one of the options) should be:

```toml
  ## We will keep connection_string commented, as we want to use environment variables
  # connection_string = ""

  ## Options between connection_string and data_format will be default and are not specified here

  ## Our data is in json, we should explicitly specify this.
  ## More information on telegraf json parser can be found here:
  ## https://github.com/influxdata/telegraf/tree/master/plugins/parsers/json
  data_format = "json"

  ## We need to parse a specific chunk of our data: records array
  ## If we don't specify this, our data will be parsed as an object with one value.
  ## What we want is to parse an array of objects
  json_query = "records"

  ## We need to explicitly specify in which field to find the name of our metric, as well as the time
  json_name_key = "metricName"
  json_time_key = "time"

  ## Explicitly specify time format (in this case it's RCF3339Nano)
  json_time_format = "2006-1-2T15:4:5.999999999Z07:00"
  json_timezone = "UTC"

  ## All string/boolean fields should be explicitly specified, otherwise they will be ignored.
  ## In our example we have 2 string fields: resourceId and timeGrain.
  ## Here we define resourceId as a tag and we omit timeGrain as we don't want this field
  tag_keys = ["resourceId"]
  ## If we wanted timeGrain to be present, but not as a tag, we could do it like this:
  # json_string_fields = ["timeGrain"]

  ## We want to add a custom tag to all the parsed metrics
  [inputs.eventhub_consumer.tags]
    azure_env = "development"
```
