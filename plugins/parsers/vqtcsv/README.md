# Vqtcsv
The vqtcsv parser creates metrics from a csv.gz file

## Configuration
```
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "vqtcsv"
```