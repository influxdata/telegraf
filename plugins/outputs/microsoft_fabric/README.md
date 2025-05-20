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
  ## The URI property of the resource on Azure
  connection_string = "https://trd-abcd.xx.kusto.fabric.microsoft.com;Database=kusto_eh; Table Name=telegraf_dump;Key=value"

  ## Client timeout
  timeout = "30s"
```

### Connection String

Fabric telegraf output plugin connection string property provide the
information necessary for a client application to establish a connection
 to a Fabric service endpoint. The connection string is a semicolon-delimited
  list of name-value parameter pairs, optionally prefixed by a single URI.

Example: data source=<https://trd-abcd.xx.kusto.fabric.microsoft.com;
Database=kusto_eh>;Table Name=telegraf_dump;Key=value

### EventHouse

The microsoft_fabric output plugin sends metrics to Microsoft Fabric,
a scalable data platform for real-time analytics.
This plugin allows you to leverage Microsoft Fabric's capabilities
to store and analyze your Telegraf metrics. Eventhouse is a high-performance,
scalable data store designed for real-time analytics. It allows you to ingest,
store, and query large volumes of data with low latency. For more information,
visit the Eventhouse documentation.

The following table lists all the possible properties that can be included in a
connection string and provide alias names for each property.

### General properties

| Property name | Description |
|---|---|
| Client Version for Tracing | The property used when tracing the client version. |
| Data Source</br></br>**Aliases:** Addr, Address, Network Address, Server | The URI specifying the Kusto service endpoint. For example, `https://mycluster.fabric.windows.net`. |
| Initial Catalog</br></br>**Alias:** Database | The default database name. For example, `MyDatabase`. |
| Ingestion Type</br></br>**Alias:** IngestionType | Values can be set   to,</br> - managed :  Streaming ingestion with fallback to batched ingestion or the "queued" method below</br> - queued :  Queue up metrics data and process sequentially |
| Table Name</br></br>**Alias:** TableName | Name of the single table to store all the metrics (Only needed if metrics_grouping_type is "SingleTable") |
| Create Tables</br></br>**Alias:** CreateTables | Creates tables and relevant mapping if set to true(default).</br>Skips table and mapping creation if set to false, this is useful for running Telegraf with the lowest possible permissions i.e. table ingestor role. |
| Metrics Grouping Type </br></br>**Alias:** MetricsGroupingType | Type of metrics grouping used when pushing to Eventhouse. values can be set, 'tablepermetric' and 'singletable'. Default is "tablepermetric" for one table per different metric.|

More about the eventhouse configuration properties
can be found [here](./EVENTHOUSE_CONFIGS.md)

### Eventstream

The eventstreams feature in the Microsoft Fabric Real-Time Intelligence
experience lets you bring real-time events into Fabric, transform them,
and then route them to various destinations without writing any code (no-code).
For more information, visit the [Eventstream documentation][eventstream_docs].  

To communicate with an eventstream, you need a connection string for the
namespace or the event hub. If you use a connection string to the namespace
from your application, following are the properties that can be added
to the standard [Eventstream connection string][ecs] like a key value pair.

| Property name | Description |
|---|---|
| Partition Key </br></br>**Aliases:**  PartitionKey | Partition key to use for the event Metric tag or field name to use for the event partition key. The value of this tag or field is set as the key for events if it exists. If both, tag and field, exist the tag is preferred. |
| Max Message Size</br></br>**Aliases:** MaxMessageSize |   Set the maximum batch message size in bytes The allowable size depends on the Event Hub tier, see <https://learn.microsoft.com/azure/event-hubs/event-hubs-quotas#basic-vs-standard-vs-premium-vs-dedicated-tiers> for details. If unset the default size defined by Azure Event Hubs is used (currently 1,000,000 bytes) |

[ecs]: https://learn.microsoft.com/azure/event-hubs/event-hubs-get-connection-string

[eventstream_docs]: https://learn.microsoft.com/fabric/real-time-intelligence/event-streams/overview?tabs=enhancedcapabilities
