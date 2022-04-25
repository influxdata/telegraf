# Google Cloud Storage Input Plugin

The Google Cloud Storage plugin will collect metrics on the given Google Cloud Storage Buckets.

## Configuration

```toml
# Read metrics from Google Cloud Storage
[[inputs.google_cloud_storage]]
  ## Required. Name of Cloud Storage bucket to ingest metrics from.
  bucket = "your-cloud-storage-bucket"

  ## Optional. Prefix of Cloud Storage bucket keys to list metrics from.
  # key_prefix = "prefix"

  ## Key that will store the offsets in order to pick up where the ingestion was left.
  offset_key = "offset_key.json"

  ## Number of objects to pull and process per iteration.
  objects_per_iteration = 10

  ## Optional. Filepath for GCP credentials JSON file to authorize calls to
  ## Google Cloud Storage APIs. If not set explicitly, Telegraf will attempt to use
  ## Application Default Credentials, which is preferred.
  credentials_file = "/Users/gkatzioura/Downloads/bigquerttest-9a75af07ee11.json"

  data_format = "json"

  json_query = "metrics"

  tag_keys = [
    "tags_datacenter",
    "tags_host"
  ]

  json_name_key = "name"

  json_time_key = "timestamp"

  json_time_format = "unix_ms"

  json_string_fields = ["cosine,sine"]

```
```