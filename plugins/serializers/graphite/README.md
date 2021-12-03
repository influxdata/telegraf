# Graphite

The Graphite data format is translated from Telegraf Metrics using either the
template pattern or tag support method.  You can select between the two
methods using the [`graphite_tag_support`](#graphite-tag-support) option.  When set, the tag support
method is used, otherwise the [Template Pattern](templates) is used.

## Configuration

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

  ## Graphite templates patterns
  ## 1. Template for cpu
  ## 2. Template for disk*
  ## 3. Default template
  # templates = [
  #  "cpu tags.measurement.host.field",
  #  "disk* measurement.field",
  #  "host.measurement.tags.field"
  #]

  ## Support Graphite tags, recommended to enable when using Graphite 1.1 or later.
  # graphite_tag_support = false
  ## Enable Graphite tags to support the full list of allowed characters
  # graphite_tag_new_sanitize = false
  ## Character for separating metric name and field for Graphite tags
  # graphite_separator = "."
```

### graphite_tag_support

When the `graphite_tag_support` option is enabled, the template pattern is not
used.  Instead, tags are encoded using
[Graphite tag support](http://graphite.readthedocs.io/en/latest/tags.html)
added in Graphite 1.1.  The `metric_path` is a combination of the optional
`prefix` option, measurement name, and field name.

The tag `name` is reserved by Graphite, any conflicting tags and will be encoded as `_name`.

**Example Conversion**:

```text
cpu,cpu=cpu-total,dc=us-east-1,host=tars usage_idle=98.09,usage_user=0.89 1455320660004257758
=>
cpu.usage_user;cpu=cpu-total;dc=us-east-1;host=tars 0.89 1455320690
cpu.usage_idle;cpu=cpu-total;dc=us-east-1;host=tars 98.09 1455320690
```

With set option `graphite_separator` to "_"

```text
cpu,cpu=cpu-total,dc=us-east-1,host=tars usage_idle=98.09,usage_user=0.89 1455320660004257758
=>
cpu_usage_user;cpu=cpu-total;dc=us-east-1;host=tars 0.89 1455320690
cpu_usage_idle;cpu=cpu-total;dc=us-east-1;host=tars 98.09 1455320690
```

The `graphite_tag_sanitize_mode` option defines how we should sanitize the tag names and values. Possible values are `strict`, or `compatible`, with the default being `strict`.

When in `strict` mode Telegraf uses the same rules as metrics when not using tags.
When in `compatible` mode Telegraf allows more characters through, and is based on the Graphite specification:
>Tag names must have a length >= 1 and may contain any ascii characters except `;!^=`. Tag values must also have a length >= 1, they may contain any ascii characters except `;` and the first character must not be `~`. UTF-8 characters may work for names and values, but they are not well tested and it is not recommended to use non-ascii characters in metric names or tags. Metric names get indexed under the special tag name, if a metric name starts with one or multiple ~ they simply get removed from the derived tag value because the ~ character is not allowed to be in the first position of the tag value. If a metric name consists of no other characters than ~, then it is considered invalid and may get dropped.

[templates]: /docs/TEMPLATE_PATTERN.md
