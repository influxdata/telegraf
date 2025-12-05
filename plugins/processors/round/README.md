# Round Processor Plugin

This plugin allows to round numerical field values to the configured precision.
This is particularly useful in combination with the [dedup processor][dedup]
to reduce the number of metrics sent to the output if only a lower precision
is required for the values.

⭐ Telegraf v1.36.0
🏷️ transformation
💻 all

[dedup]: /plugins/processors/dedup/README.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Round numerical fields
[[processors.round]]
  ## Precision to round to.
  ## A positive number indicates rounding to the right of the decimal separator (the fractional part).
  ## A negative number indicates rounding to the left of the decimal separator.
  # precision = 0

  ## Round only numeric fields matching the filter criteria below.
  ## Excludes takes precedence over includes.
  # include_fields = ["*"]
  # exclude_fields = []
```

## Example

Round each value the _inputs.cpu_ plugin generates, except for the
`usage_steal`, `usage_user`, `uptime_format`, `usage_idle` field:

```toml
[[inputs.cpu]]
  percpu = true
  totalcpu = true
  collect_cpu_time = false
  report_active = false

[[processors.round]]
  precision = 1
  include_fields = []
  exclude_fields = ["usage_steal", "usage_user", "uptime_format", "usage_idle" ]
```

Result of rounding the _cpu_ metric:

```diff
- cpu map[cpu:cpu11 host:98d5b8dbad1c] map[usage_guest:0 usage_guest_nice:0 usage_idle:94.3999999994412 usage_iowait:0 usage_irq:0.1999999999998181 usage_nice:0 usage_softirq:0.20000000000209184 usage_steal:0 usage_system:1.2000000000080036 usage_user:4.000000000014552]
+ cpu map[cpu:cpu11 host:98d5b8dbad1c] map[usage_guest:0 usage_guest_nice:0 usage_idle:94.4 usage_iowait:0 usage_irq:0.2 usage_nice:0 usage_softirq:0.2 usage_steal:0 usage_system:1.2 usage_user:4.0]
```
