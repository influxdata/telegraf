# Microsoft Fabric Output Plugin

This plugin writes metrics to [Fabric Eventhouse][eventhouse] and
[Fabric Eventstream][eventstream] artifacts of
 [Real-Time Intelligence in Microsoft Fabric][fabric].

Real-Time Intelligence is a SaaS service in Microsoft Fabric
that allows you to extract insights and visualize data in motion.
It offers an end-to-end solution for event-driven scenarios,
 streaming data, and data logs.

⭐ Telegraf v1.35.0
🏷️ datastore
💻 all

[eventhouse]:
 https://learn.microsoft.com/fabric/real-time-intelligence/eventhouse
[eventstream]:
 https://learn.microsoft.com/fabric/real-time-intelligence/event-streams/overview?tabs=enhancedcapabilities
[fabric]:
 https://learn.microsoft.com/fabric/real-time-intelligence/overview

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Sends metrics to Microsoft Fabric
[[outputs.microsoft_fabric]]
  ## The URI property of the resource on Microsoft Fabric
  connection_string = "https://trd-abcd.xx.kusto.fabric.microsoft.com;Database=kusto_eh;Table Name=telegraf_dump;Key=value"

  ## Client timeout
  # timeout = "30s"
```

### Connection String

The `connection_string` provide information necessary for the plugin to
establish a connection to the Fabric service endpoint. It is a
semicolon-delimited list of name-value parameter pairs, optionally prefixed by
a single URI. The setting is specific to the type of endpoint you are using.
The sections below will detail on the required and available name-value pairs
for each type.

### EventHouse

This plugin allows you to leverage Microsoft Fabric's capabilities to store and
analyze your Telegraf metrics. Eventhouse is a high-performance, scalable
data-store designed for real-time analytics. It allows you to ingest, store and
query large volumes of data with low latency. For more information, visit the
[Eventhouse documentation][eventhousedocs].

The following table lists all the possible properties that can be included in a
connection string and provide alias names for each property.

| Property name | Aliases | Description |
|---|---|---|
| Client Version for Tracing | | The property used when tracing the client version. |
| Data Source | Addr, Address, Network Address, Server | The URI specifying the Eventhouse service endpoint. For example, `https://mycluster.fabric.windows.net`. |
| Initial Catalog | Database | The default database name. For example, `MyDatabase`. |
| Ingestion Type | IngestionType | Values can be set to `managed` for streaming ingestion with fallback to batched ingestion or the `queued` method for queuing up metrics and process sequentially |
| Table Name | TableName | Name of the single table to store all the metrics; only needed if `metrics_grouping_type` is `singletable` |
| Create Tables | CreateTables | Creates tables and relevant mapping if `true` (default). Otherwise table and mapping creation is skipped. This is useful for running Telegraf with the lowest possible permissions i.e. table ingestor role. |
| Metrics Grouping Type | MetricsGroupingType | Type of metrics grouping used when pushing to Eventhouse either being `tablepermetric` or `singletable`. Default is "tablepermetric" for one table per different metric.|

[eventhousedocs]: https://learn.microsoft.com/fabric/real-time-intelligence/eventhouse

#### Metrics Grouping

Metrics can be grouped in two ways to be sent to Eventhouse. To specify
which metric grouping type the plugin should use, the respective value should be
given to the `Metrics Grouping Type` in the connection string. If no value is
given, by default, the metrics will be grouped using `tablepermetric`.

#### TablePerMetric

The plugin will group the metrics by the metric name and will send each group
of metrics to an Eventhouse KQL DB table. If the table doesn't exist the
plugin will create the table, if the table exists then the plugin will try to
merge the Telegraf metric schema to the existing table. For more information
about the merge process check the [`.create-merge` documentation][create-merge].

The table name will match the metric name, i.e. the name of the metric must
comply with the Eventhouse KQL DB table naming constraints in case you plan
to add a prefix to the metric name.

[create-merge]: https://learn.microsoft.com/kusto/management/create-merge-tables-command?view=microsoft-fabric

#### SingleTable

The plugin will send all the metrics received to a single Eventhouse KQL DB
table. The name of the table must be supplied via `table_name` parameter in the
`connection_string`. If the table doesn't exist the plugin will create the
table, if the table exists then the plugin will try to merge the Telegraf metric
schema to the existing table. For more information about the merge process check
the [`.create-merge` documentation][create-merge].

#### Tables Schema

The schema of the Eventhouse table will match the structure of the metric.
The corresponding command generated by the plugin would be like the following:

```kql
.create-merge table ['table-name']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime)
```

The corresponding table mapping would be like the following:

```kql
.create-or-alter table ['table-name'] ingestion json mapping 'table-name_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'
```

> [!NOTE]
> This plugin will automatically create tables and corresponding table mapping
> using the command above.

#### Ingestion type

> [!NOTE]
> [Streaming ingestion][streaming] has to be enabled on Eventhouse in case of
> `managed` operation.

Refer to the following query below to check if streaming is enabled:

```kql
.show database <DB-Name> policy streamingingestion
```

To learn more about configuration, supported authentication methods and querying
ingested data, check the [documentation][ethdocs].

[streaming]: https://learn.microsoft.com/azure/data-explorer/ingest-data-streaming?tabs=azure-portal%2Ccsharp
[ethdocs]: https://learn.microsoft.com/azure/data-explorer/ingest-data-telegraf

### Eventstream

Eventstreams allow you to bring real-time events into Fabric, transform them,
and then route them to various destinations without writing any code (no-code).
For more information, visit the [Eventstream documentation][eventstream_docs].

To communicate with an eventstream, you need to specify a connection string for
the namespace or the event hub. The following properties can be added to the
standard [Eventstream connection string][ecs] using key-value pairs:

| Property name | Aliases | Description |
|---|---|---|
| Partition Key | PartitionKey | Metric tag or field name to use for the event partition key if it exists. If both, tag and field, exist the tag is takes precedence, otherwise the value `<default>` is used |
| Max Message Size| MaxMessageSize | Maximum batch message size in bytes The allowable size depends on the Event Hub tier, see [tier information][tiers] for details. If unset the default size defined by Eventstream is used (currently 1,000,000 bytes) |

[eventstream_docs]: https://learn.microsoft.com/fabric/real-time-intelligence/event-streams/overview?tabs=enhancedcapabilities
[ecs]: https://learn.microsoft.com/azure/event-hubs/event-hubs-get-connection-string
[tiers]: https://learn.microsoft.com/azure/event-hubs/event-hubs-quotas#basic-vs-standard-vs-premium-vs-dedicated-tiers
