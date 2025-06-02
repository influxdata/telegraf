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

The `connection_string` provide the information necessary for a client application to
establish a connection to the Fabric service endpoint. It is a
semicolon-delimited list of name-value parameter pairs, optionally prefixed by a
single URI. The `connection_string` setting is specific to the type of endpoint you are using.
The sections below will detail on the required and available name-value pairs for
each type.

### EventHouse

This plugin allows you to leverage Microsoft Fabric's capabilities
to store and analyze your Telegraf metrics. Eventhouse is a high-performance,
scalable data store designed for real-time analytics. It allows you to ingest,
store, and query large volumes of data with low latency. For more information,
visit the [Eventhouse documentation][eventhousedocs].

[eventhousedocs]: https://learn.microsoft.com/fabric/real-time-intelligence/eventhouse

The following table lists all the possible properties that can be included in a
connection string and provide alias names for each property:

The following table lists all the possible properties that can be included in a
connection string and provide alias names for each property.

| Property name | Description |
|---|---|
| Client Version for Tracing | The property used when tracing the client version. |
| Data Source</br></br>**Aliases:** Addr, Address, Network Address, Server | The URI specifying the Kusto service endpoint. For example, `https://mycluster.fabric.windows.net`. |
| Initial Catalog</br></br>**Alias:** Database | The default database name. For example, `MyDatabase`. |
| Ingestion Type</br></br>**Alias:** IngestionType | Values can be set   to,</br> - managed :  Streaming ingestion with fallback to batched ingestion or the "queued" method below</br> - queued :  Queue up metrics data and process sequentially |
| Table Name</br></br>**Alias:** TableName | Name of the single table to store all the metrics (Only needed if metrics_grouping_type is "SingleTable") |
| Create Tables</br></br>**Alias:** CreateTables | Creates tables and relevant mapping if set to true(default).</br>Skips table and mapping creation if set to false, this is useful for running Telegraf with the lowest possible permissions i.e. table ingestor role. |
| Metrics Grouping Type </br></br>**Alias:** MetricsGroupingType | Type of metrics grouping used when pushing to Eventhouse. values can be set, 'tablepermetric' and 'singletable'. Default is "tablepermetric" for one table per different metric.|

* *Metrics Grouping*

  Metrics can be grouped in two ways to be sent to Azure Data Explorer. To specify
  which metric grouping type the plugin should use, the respective value should be
  given to the `Metrics Grouping Type` in the connection string. If no value is given, by default, the metrics will be grouped using
  `tablepermetric`.

* *TablePerMetric*

  The plugin will group the metrics by the metric name, and will send each group
  of metrics to an Azure Data Explorer table. If the table doesn't exist the
  plugin will create the table, if the table exists then the plugin will try to
  merge the Telegraf metric schema to the existing table. For more information
about the merge process check the [`.create-merge` documentation][create-merge].

  The table name will match the `name` property of the metric, this means that the
  name of the metric should comply with the Azure Data Explorer table naming
  constraints in case you plan to add a prefix to the metric name.

[create-merge]: https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/create-merge-table-command

* *SingleTable*

  The plugin will send all the metrics received to a single Azure Data Explorer
  table. The name of the table must be supplied via `table_name` in the config
  file. If the table doesn't exist the plugin will create the table, if the table
  exists then the plugin will try to merge the Telegraf metric schema to the
  existing table. For more information about the merge process check the
  [`.create-merge` documentation][create-merge].

* *Tables Schema*

  The schema of the Eventhouse table will match the structure of the
  Telegraf `Metric` object. The corresponding command
  generated by the plugin would be like the following:

  ```kql
  .create-merge table ['table-name']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime)
  ```

  The corresponding table mapping would be like the following:

  ```kql
  .create-or-alter table ['table-name'] ingestion json mapping 'table-name_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'
  ```

  **Note**: This plugin will automatically create Eventhouse tables and
  corresponding table mapping as per the above mentioned commands.

* *Ingestion type*

  **Note**:
  [Streaming ingestion](https://aka.ms/AAhlg6s)
  has to be enabled on ADX [configure the ADX cluster]
  in case of `managed` option.
  Refer the query below to check if streaming is enabled

  ```kql
  .show database <DB-Name> policy streamingingestion
  ```

  To know more about configuration, supported authentication methods and querying ingested data, read the [documentation][ethdocs]

  [ethdocs]: https://learn.microsoft.com/azure/data-explorer/ingest-data-telegraf

### Eventstream

Eventstreams allow you to bring real-time events into Fabric, transform them,
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
