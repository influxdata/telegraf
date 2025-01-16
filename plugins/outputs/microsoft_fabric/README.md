# Microsoft Fabric Output Plugin

This plugin writes metrics to [Real time analytics in Fabric][fabric] services.

[fabric]: https://learn.microsoft.com/en-us/fabric/real-time-analytics/overview

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml
# Sends metrics to Microsoft Fabric
[[outputs.microsoft_fabric]]
  ## The URI property of the resource on Azure
  ## ex: connection_string = "https://myadxresource.australiasoutheast.kusto.windows.net"
  ## ex: connection_string = "Endpoint=sb://namespace.servicebus.windows.net/;*****;EntityPath=hubName"
  connection_string = ""


    [outputs.microsoft_fabric.eh_conf]
        ## The Eventhouse database that the metrics will be ingested into.
        ## The plugin will NOT generate this database automatically, it's expected that this database already exists before ingestion.
        ## ex: "exampledatabase"
        database = ""

        ## Timeout for Eventhouse operations
        # timeout = "20s"

        ## Type of metrics grouping used when pushing to Eventhouse.
        ## Default is "TablePerMetric" for one table per different metric.
        ## For more information, please check the plugin README.
        # metrics_grouping_type = "TablePerMetric"

        ## Name of the single table to store all the metrics (Only needed if metrics_grouping_type is "SingleTable").
        # table_name = ""

        ## Creates tables and relevant mapping if set to true(default).
        ## Skips table and mapping creation if set to false, this is useful for running Telegraf with the lowest possible permissions i.e. table ingestor role.
        # create_tables = true

        ##  Ingestion method to use.
        ##  Available options are
        ##    - managed  --  streaming ingestion with fallback to batched ingestion or the "queued" method below
        ##    - queued   --  queue up metrics data and process sequentially
        # ingestion_type = "queued"

    [outputs.microsoft_fabric.es_conf]
        ## The full connection string to the Event stream (required)
        ## The shared access key must have "Send" permissions on the target Event stream.
        
        ## Client timeout (defaults to 30s)
        # timeout = "30s"

        ## Partition key
        ## Metric tag or field name to use for the event partition key. The value of
        ## this tag or field is set as the key for events if it exists. If both, tag
        ## and field, exist the tag is preferred.
        # partition_key = ""

        ## Set the maximum batch message size in bytes
        ## The allowable size depends on the Event stream tier
        ## See: https://learn.microsoft.com/azure/event-hubs/event-hubs-quotas#basic-vs-standard-vs-premium-vs-dedicated-tiers
        ## Setting this to 0 means using the default size from the Azure Event streams Client library (1000000 bytes)
        # max_message_size = 1000000

        ## Data format to output.
        ## Each data format has its own unique set of configuration options, read
        ## more about them here:
        ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
        data_format = "json"
```

## Description

The `microsoft_fabric` output plugin sends metrics to Microsoft Fabric,
a scalable data platform for real-time analytics.
This plugin allows you to leverage Microsoft Fabric's
capabilities to store and analyze your Telegraf metrics.
Following are the currently supported datastores:

### Eventhouse

Eventhouse is a high-performance, scalable data store designed for
 real-time analytics. It allows you to ingest, store, and query large
 volumes of data with low latency.  For more information, visit the
 [Eventhouse documentation](
    https://learn.microsoft.com/fabric/real-time-intelligence/eventhouse
    ).

```toml
[outputs.microsoft_fabric.eh_conf]
        ## The Eventhouse database that the metrics will be ingested into.
        ## The plugin will NOT generate this database automatically, it's expected that this database already exists before ingestion.
        ## ex: "exampledatabase"
        database = ""

        ## Timeout for Eventhouse operations
        # timeout = "20s"

        ## Type of metrics grouping used when pushing to Eventhouse.
        ## Default is "TablePerMetric" for one table per different metric.
        ## For more information, please check the plugin README.
        # metrics_grouping_type = "TablePerMetric"

        ## Name of the single table to store all the metrics (Only needed if metrics_grouping_type is "SingleTable").
        # table_name = ""

        ## Creates tables and relevant mapping if set to true(default).
        ## Skips table and mapping creation if set to false, this is useful for running Telegraf with the lowest possible permissions i.e. table ingestor role.
        # create_tables = true

        ##  Ingestion method to use.
        ##  Available options are
        ##    - managed  --  streaming ingestion with fallback to batched ingestion or the "queued" method below
        ##    - queued   --  queue up metrics data and process sequentially
        # ingestion_type = "queued"

```

More about the eventhouse configuration properties
can be found [here](../azure_data_explorer/README.md#metrics-grouping)

### Eventstream

The eventstreams feature in the Microsoft Fabric Real-Time Intelligence
experience lets you bring real-time events into Fabric, transform them,
and then route them to various destinations without writing any code (no-code).
For more information, visit the [Eventstream documentation][].

[Eventstream documentation]: https://learn.microsoft.com/fabric/real-time-intelligence/event-streams/overview?tabs=enhancedcapabilities 

```toml
[outputs.microsoft_fabric.es_conf]
        ## The full connection string to the Event stream (required)
        ## The shared access key must have "Send" permissions on the target Event stream.
        
        ## Client timeout (defaults to 30s)
        # timeout = "30s"

        ## Partition key
        ## Metric tag or field name to use for the event partition key. The value of
        ## this tag or field is set as the key for events if it exists. If both, tag
        ## and field, exist the tag is preferred.
        # partition_key = ""

        ## Set the maximum batch message size in bytes
        ## The allowable size depends on the Event stream tier
        ## See: https://learn.microsoft.com/azure/event-hubs/event-hubs-quotas#basic-vs-standard-vs-premium-vs-dedicated-tiers
        ## Setting this to 0 means using the default size from the Azure Event streams Client library (1000000 bytes)
        # max_message_size = 1000000

        ## Data format to output.
        ## Each data format has its own unique set of configuration options, read
        ## more about them here:
        ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
        data_format = "json"

```
