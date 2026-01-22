# Google BigQuery Output Plugin

This plugin writes metrics to the [Google Cloud BigQuery][big_query] service
and requires [authentication][authentication] with Google Cloud using either a
service account or user credentials.

> [!IMPORTANT]
> Be aware that this plugin accesses APIs that are [chargeable][pricing] and
> might incur costs.

[authentication]: https://cloud.google.com/bigquery/docs/authentication
[big_query]: https://cloud.google.com/bigquery
[pricing]: https://cloud.google.com/bigquery/pricing

⭐ Telegraf v1.18.0
🏷️ cloud, datastore
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Configuration for Google Cloud BigQuery to send entries
[[outputs.bigquery]]
  ## Credentials File
  credentials_file = "/path/to/service/account/key.json"

  ## Google Cloud Platform Project
  # project = ""

  ## The namespace for the metric descriptor
  dataset = "telegraf"

  ## Timeout for BigQuery operations.
  # timeout = "5s"

  ## Character to replace hyphens on Metric name
  # replace_hyphen_to = "_"

  ## Write all metrics in a single compact table
  # compact_table = ""
```

Leaving `project` empty indicates the plugin will try to retrieve the project
from the credentials file.

Requires `dataset` to specify under which BigQuery dataset the corresponding
metrics tables reside.

Each metric should have a corresponding table to BigQuery.  The schema of the
table on BigQuery:

* Should contain the field `timestamp` which is the timestamp of a telegraph
  metrics
* Should contain the metric's tags with the same name and the column type should
  be set to string.
* Should contain the metric's fields with the same name and the column type
  should match the field type.

## Compact table

When enabling the compact table, all metrics are inserted to the given table
with the following schema:

```json
[
  {
    "mode": "REQUIRED",
    "name": "timestamp",
    "type": "TIMESTAMP"
  },
  {
    "mode": "REQUIRED",
    "name": "name",
    "type": "STRING"
  },
  {
    "mode": "REQUIRED",
    "name": "tags",
    "type": "JSON"
  },
  {
    "mode": "REQUIRED",
    "name": "fields",
    "type": "JSON"
  }
]
```

## Restrictions

Avoid hyphens on BigQuery tables, underlying SDK cannot handle streaming inserts
to Table with hyphens.

In cases of metrics with hyphens please use the [Rename Processor
Plugin][rename].

In case of a metric with hyphen by default hyphens shall be replaced with
underscores (_).  This can be altered using the `replace_hyphen_to`
configuration property.

Available data type options are:

* integer
* float or long
* string
* boolean

All field naming restrictions that apply to BigQuery should apply to the
measurements to be imported.

Tables on BigQuery should be created beforehand and they are not created during
persistence

Pay attention to the column `timestamp` since it is reserved upfront and cannot
change.  If partitioning is required make sure it is applied beforehand.

[rename]: ../../processors/rename/README.md
