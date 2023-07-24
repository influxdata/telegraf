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

### Batch mode

When an output plugin emits multiple metrics in a batch fashion, by default the
template will just be repeated for each metric. If you would like to specifically
define how a batch should be formatted, you can use a `batch_template` instead.
In this mode, the context of the template (the 'dot') will be a slice of metrics.

```toml
batch_template = '''My batch metric names: {{range $index, $metric := . -}}
{{if $index}}, {{ end }}{{ $metric.Name }}
{{- end }}'''
```
