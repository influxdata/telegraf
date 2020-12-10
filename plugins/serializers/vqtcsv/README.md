# Vqtcsv Serializer
This plugin parses csv.gz files and creates custom output of the metrics

## Configuration
```
[[outputs.file]]
  files = [/path/to/destination]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "vqtcsv"  
```