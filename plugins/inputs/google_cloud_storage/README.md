# Google Cloud Storage Input Plugin

The Google Cloud Storage plugin will collect metrics
on the given Google Cloud Storage Buckets.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Gather metrics by iterating the files located on a Cloud Storage Bucket.
[[inputs.google_cloud_storage]]
  ## Required. Name of Cloud Storage bucket to ingest metrics from.
  bucket = "my-bucket"

  ## Optional. Prefix of Cloud Storage bucket keys to list metrics from.
  # key_prefix = "my-bucket"

  ## Key that will store the offsets in order to pick up where the ingestion was left.
  offset_key = "offset_key"

  ## Key that will store the offsets in order to pick up where the ingestion was left.
  objects_per_iteration = 10

  ## Required. Data format to consume.
  ## Each data format has its own unique set of configuration options.
  ## Read more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
  
  ## Optional. Filepath for GCP credentials JSON file to authorize calls to
  ## Google Cloud Storage APIs. If not set explicitly, Telegraf will attempt to use
  ## Application Default Credentials, which is preferred.
  # credentials_file = "path/to/my/creds.json"
```

## Metrics

- Measurements will reside on Google Cloud Storage with the format specified

- example when [[inputs.google_cloud_storage.data_format]] is json

```json
{
  "metrics": [
    {
      "fields": {
        "cosine": 10,
        "sine": -1.0975806427415925e-12
      },
      "name": "cpu",
      "tags": {
        "datacenter": "us-east-1",
        "host": "localhost"
      },
      "timestamp": 1604148850990
    }
  ]
}
```

## Example Output

The example output

```shell
2022-09-17T17:52:19Z I! Starting Telegraf 1.25.0-a93ec9a0
2022-09-17T17:52:19Z I! Available plugins: 209 inputs, 9 aggregators, 26 processors, 20 parsers, 57 outputs
2022-09-17T17:52:19Z I! Loaded inputs: google_cloud_storage
2022-09-17T17:52:19Z I! Loaded aggregators: 
2022-09-17T17:52:19Z I! Loaded processors: 
2022-09-17T17:52:19Z I! Loaded outputs: influxdb
2022-09-17T17:52:19Z I! Tags enabled: host=user-N9RXNWKWY3
2022-09-17T17:52:19Z I! [agent] Config: Interval:10s, Quiet:false, Hostname:"user-N9RXNWKWY3", Flush Interval:10s
```
