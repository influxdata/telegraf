# Template Serializer

The `template` output data format outputs metrics using an user defined go template.

## Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "template"

  ## Go template which defines output format
  template = '{{ .Tag "host" }} {{ .Field "available" }}'
```
