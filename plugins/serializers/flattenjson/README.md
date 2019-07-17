# Flatten Json serializer

The Flatten Json serializer allow you to flatten json. The specificity is that each index fields value will give one output and each output contains all the tags (prefixed by "tags_"), metric_family, metric_name, metric_value, timestamp.


As an example, the following event:
```javascript
{
  "fields": {
    "reads": 973,
    "write_bytes": 2097152
  },
  "name": "diskio",
  "tags": {
    "host": "myhost",
    "name": "vda1"
  },
  "timestamp": 1556936700000
}
```

Will be flatten by flattenjson plugin as below:
```javascript
{
  'metric_family': 'diskio',
  'timestamp': 1556936700000,
  'tags_host': 'myhost',
  'tags_name': 'vda1',
  'metric_name': 'reads',
  'metric_value': 973
}
{
  'metric_family': 'diskio',
  'timestamp': 1556936700000,
  'tags_host': 'myhost',
  'tags_name': 'vda1',
  'metric_name': 'write_bytes',
  'metric_value': 2097152
}
```

## Using with the File output

An example configuration of a file based output is:

```toml
 # Send telegraf metrics to file(s)
[[outputs.file]]
   ## Files to write to, "stdout" is a specially handled file.
   files = ["/tmp/metrics.out"]

   ## Data format to output.
   ## Each data format has its own unique set of configuration options, read
   ## more about them here:
   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
   data_format = "flattenjson"
```
