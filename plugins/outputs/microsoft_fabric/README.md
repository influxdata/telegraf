# Microsoft Fabric Output Plugin

This plugin writes metrics to [Real time analytics in Fabric][fabric] services.

⭐ Telegraf v1.35.0
🏷️ datastore
💻 all

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
  ## The URI property of the Eventhouse resource on Azure
  ## ex: connection_string = "Data Source=https://myadxresource.australiasoutheast.kusto.windows.net"
  connection_string = ""


  ## Using this section the plugin will send metrics to an Eventhouse endpoint
  ## for ingesting, storing, and querying large  volumes of data with low latency.
  [outputs.microsoft_fabric.eventhouse]
    [outputs.microsoft_fabric.eventhouse.cluster_config]
        ## Database metrics will be written to  
      ## NOTE: The plugin will NOT generate the database. It is expected the database already exists.  
      database = ""

      ## Timeout for Eventhouse operations
      # timeout = "20s"

      ## Type of metrics grouping; available options are:
      ##   tablepermetric -- for one table per distinct metric
      ##   singletable    -- for writing all metrics to the same table
      # metrics_grouping_type = "tablepermetric"

      # Name of the table to store metrics
      ## NOTE: This option is only used for "singletable" metrics grouping
      # table_name = ""

      ## Creates tables and relevant mapping
      ## Disable when running with the lowest possible permissions i.e. table ingestor role.
      # create_tables = true

      ##  Ingestion method to use; available options are
      ##    - managed  --  streaming ingestion with fallback to batched ingestion or the "queued" method below
      ##    - queued   --  queue up metrics data and process sequentially
      # ingestion_type = "queued"

  ## Using this section the plugin will send metrics to an EventStream endpoint  
  ## for transforming and routing metrics to various destinations without writing  
  ## any code.  
  [outputs.microsoft_fabric.eventstream] 
    ## The full connection string to the Event Hub (required)
    ## The shared access key must have "Send" permissions on the target Event Hub.
    
    ## Client timeout
    # timeout = "30s"

    ## Partition key
    ## Metric tag or field name to use for the event partition key. The value of
    ## this tag or field is set as the key for events if it exists. If both, tag
    ## and field, exist the tag is preferred.
    # partition_key = ""

    ## Set the maximum batch message size in bytes  
    ## The allowable size depends on the Event Hub tier; not setting this option or setting  
    ## it to zero will use the default size of the Azure Event Hubs Client library. See  
    ##   https://learn.microsoft.com/azure/event-hubs/event-hubs-quotas#basic-vs-standard-vs-premium-vs-dedicated-tiers  
    ## for the allowable size of your tier.  
    # max_message_size = "0B"  

    ## Data format to output.
    ## Each data format has its own unique set of configuration options, read
    ## more about them here:
    ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
    data_format = "json"
```

### EventHouse

The microsoft_fabric output plugin sends metrics to Microsoft Fabric, a scalable data platform for real-time analytics. This plugin allows you to leverage Microsoft Fabric's capabilities to store and analyze your Telegraf metrics. Eventhouse is a high-performance, scalable data store designed for real-time analytics. It allows you to ingest, store, and query large volumes of data with low latency. For more information, visit the Eventhouse documentation.


More about the eventhouse configuration properties
can be found [here](./EVENTHOUSE_CONFIGS.md)

### Eventstream

The eventstreams feature in the Microsoft Fabric Real-Time Intelligence
experience lets you bring real-time events into Fabric, transform them,
and then route them to various destinations without writing any code (no-code).
For more information, visit the [Eventstream documentation][eventstream_docs].  

[eventstream_docs]: https://learn.microsoft.com/fabric/real-time-intelligence/event-streams/overview?tabs=enhancedcapabilities  
