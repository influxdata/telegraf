# JSON template serializer

The `json_template` output data format converts metrics into JSON documents using a custom [template][golang_templates].

## Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "json_template"

  ## The template to use for the output using Golang templates.
  ## See https://pkg.go.dev/text/template for details.
  template = '''
 	[
{{- range $idx, $metric := . -}}
		{
			"name":  "{{ $metric.Name }}",
			"fields": {
				{{ template "fields" $metric }}
			},
			"tags": {
				{{ template "tags" $metric }}
			},
			"timestamp":   {{ $metric.Time.Unix }}
		}{{- if not (last $idx $metrics)}},{{- end }}
{{- end -}}
	]
  '''

  ## The output style for the resulting JSON.
  ## Possible values are
  ##  "raw"      unmodified text produced by the template (default)
  ##  "compact"  compact form with all unnecessary spaces removed,
  ##             saved bandwidth
  ##  "pretty"   nicely (two-space) indented JSON, good for humans
  #json_template_style = "raw"
```

The above configuration reproduces the output of the `json` serializer.

## Available functions, templates and variables

To ease template definition, the following additional functions are defined

- `last(index, array/slice)`: returns `true` for `index` addressing the last
                              element in the given array or slice
- `uppercase`: returns the uppercased version of a string
- `lowercase`: returns the lowercased version of a string

Furthermore, the following templates are defined

- `fields`: outputs the fields of a given metric in key-value form with correct quoting strings
- `tags`: outputs the tags of a given metric in key-value form with correct quoting

The following varibles are globally defined

- `$metrics`: reference to the metrics array in batch mode and to the single metric in
              non-batch mode. This might be helpful when iterating through the metrics
              and an absolute  reference to one of the metrics is required.

## Examples

The above template produces a JSON in the form:

```json
{
    "fields": {
        "field_1": 30,
        "field_2": 4,
        "field_N": 59,
        "n_images": 660
    },
    "name": "docker",
    "tags": {
        "host": "raynor"
    },
    "timestamp": 1458229140
}
```

When an output plugin needs to emit multiple metrics at one time, it may use
the batch format. The use of batch format is determined by the plugin,
reference the documentation for the specific plugin.

[golang_templates]: https://pkg.go.dev/text/template