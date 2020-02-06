# Template Processor

The `template` processor applies a Go template to metrics to generate a new
tag.  The primary use case of this plugin is to create a tag that can be used
for dynamic routing to multiple output plugins or using an output specific
routing option.

The template has access to each metric's measurement name, tags, fields, and
timestamp using the [interface in `/template_metric.go`](template_metric.go).

Read the full [Go Template Documentation][].

### Configuration

```toml
[[processors.template]]
  ## Tag to set with the output of the template.
  tag = "topic"

  ## Go template used to create the tag value.  In order to ease TOML
  ## escaping requirements, you may wish to use single quotes around the
  ## template string.
  template = '{{ .Tag "hostname" }}.{{ .Tag "level" }}'
```

### Example

Combine multiple tags to create a single tag:
```toml
[[processors.template]]
  tag = "topic"
  template = '{{ .Tag "hostname" }}.{{ .Tag "level" }}'
```

```diff
- cpu,level=debug,hostname=localhost time_idle=42
+ cpu,level=debug,hostname=localhost,topic=localhost.debug time_idle=42
```

Add measurement name as a tag:
```toml
[[processors.template]]
  tag = "measurement"
  template = '{{ .Name }}'
```

```diff
- cpu,hostname=localhost time_idle=42
+ cpu,hostname=localhost,meaurement=cpu time_idle=42
```

Add the year as a tag, similar to the date processor:
```toml
[[processors.template]]
  tag = "year"
  template = '{{.Time.UTC.Year}}'
```

[Go Template Documentation]: https://golang.org/pkg/text/template/
