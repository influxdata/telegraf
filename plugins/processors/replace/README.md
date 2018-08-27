# Replace Processor Plugin

The `replace` processor runs a `strings.replace` on metric names and replaces
specified chars with replacement chars. Multiple replacement configs can be
stacked on top of each other and executed sequentially. This is useful when
an input module uses a different set of metric separator than the output module,
or if there are parts of the metric name that are not useful.

Note that if the char deletion aspect of this is used, and the entire name is
deleted, it will not perform the operation and the original name will be left
as is.

### Configuration:

The following example shows a replacement function where it first replace all
`_` with `-`, then we replace all `:` with `_`. For example if an incoming
metric has the name `average:cpu:usage_percentage`, then it will exit with the
name `average_cpu_usage-percentage`.

```toml
[[processors.replace]]
  old = "_"
  new = "-"

[[processors.replace]]
  old = ":"
  new = "_"
```

