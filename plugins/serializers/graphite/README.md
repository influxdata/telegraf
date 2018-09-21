# Graphite

The Graphite data format is translated from Telegraf Metrics using either the
template pattern or tag support method.  You can select between the two
methods using the [`graphite_tag_support`](#graphite-tag-support) option.  When set, the tag support
method is used, otherwise the [Template Pattern](templates) is used.

### Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "graphite"

  ## Prefix added to each graphite bucket
  prefix = "telegraf"
  ## Graphite template pattern
  template = "host.tags.measurement.field"

  ## Support Graphite tags, recommended to enable when using Graphite 1.1 or later.
  # graphite_tag_support = false
```

#### graphite_tag_support

When the `graphite_tag_support` option is enabled, the template pattern is not
used.  Instead, tags are encoded using
[Graphite tag support](http://graphite.readthedocs.io/en/latest/tags.html)
added in Graphite 1.1.  The `metric_path` is a combination of the optional
`prefix` option, measurement name, and field name.

The tag `name` is reserved by Graphite, any conflicting tags and will be encoded as `_name`.

**Example Conversion**:
```
cpu,cpu=cpu-total,dc=us-east-1,host=tars usage_idle=98.09,usage_user=0.89 1455320660004257758
=>
cpu.usage_user;cpu=cpu-total;dc=us-east-1;host=tars 0.89 1455320690
cpu.usage_idle;cpu=cpu-total;dc=us-east-1;host=tars 98.09 1455320690
```

[templates]: /docs/TEMPLATE_PATTERN.md
