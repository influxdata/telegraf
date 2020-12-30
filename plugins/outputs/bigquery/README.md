# BigQuery Google Cloud Output Plugin

This plugin writes to the [Google Cloud BigQuery][bigquery] and requires [authentication][] 
with Google Cloud using either a service account or user credentials

This plugin accesses APIs which are [chargeable][pricing]; you might incur
costs.

Requires `project` to specify where BigQuery entries will be persisted.

Requires `dataset` to specify under which BigQuery dataset the corresponding metrics tables reside.

Each metric should have a corresponding table to BigQuery. 
The schema of the table on BigQuery:
* Should contain the field `timestamp` which is the timestamp of a telegraph metrics
* Should contain the metric's tags with the same name and the column type should be set to string.
* Should contain the metric's fields with the same name and the column type should match the field type.

### Configuration

```toml
[[outputs.bigquery]]
  ## GCP Project
  project = "erudite-bloom-151019"

  ## The BigQuery dataset
  dataset = "telegraf"
```

### Restrictions

Current sdk cannot handle inserts to Table with hyphens.

Available data type options are:
* integer
* float or long
* string
* boolean

All field naming restrictions that apply to BigQuery sould apply to the measurements to be imported.

Tables on BigQuery should be created before hand and they are not created during persistence

Pay attention to the columnd `timestamp` since it is reserved upfront and cannot change. 
If partitioning is required make sure it is applied beforehand.

