# Prometheus

The `prometheus` data format converts metrics into the Prometheus text
exposition format.  When used with the `prometheus` input, the input should be
use the `metric_version = 2` option in order to properly round trip metrics.

**Warning**: When generating histogram and summary types, output may
not be correct if the metric spans multiple batches.  This issue can be
somewhat, but not fully, mitigated by using outputs that support writing in
"batch format".  When using histogram and summary types, it is recommended to
use only the `prometheus_client` output.

## Configuration

```toml
[[outputs.file]]
  files = ["stdout"]
  use_batch_format = true

  ## Include the metric timestamp on each sample.
  prometheus_export_timestamp = false

  ## Sort prometheus metric families and metric samples.  Useful for
  ## debugging.
  prometheus_sort_metrics = false

  ## Output string fields as metric labels; when false string fields are
  ## discarded.
  prometheus_string_as_label = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "prometheus"
```

### Example

**Example Input**
```
cpu,cpu=cpu0 time_guest=8022.6,time_system=26145.98,time_user=92512.89 1574317740000000000
cpu,cpu=cpu1 time_guest=8097.88,time_system=25223.35,time_user=96519.58 1574317740000000000
cpu,cpu=cpu2 time_guest=7386.28,time_system=24870.37,time_user=95631.59 1574317740000000000
cpu,cpu=cpu3 time_guest=7434.19,time_system=24843.71,time_user=93753.88 1574317740000000000
```

**Example Output**
```
# HELP cpu_time_guest Telegraf collected metric
# TYPE cpu_time_guest counter
cpu_time_guest{cpu="cpu0"} 9582.54
cpu_time_guest{cpu="cpu1"} 9660.88
cpu_time_guest{cpu="cpu2"} 8946.45
cpu_time_guest{cpu="cpu3"} 9002.31
# HELP cpu_time_system Telegraf collected metric
# TYPE cpu_time_system counter
cpu_time_system{cpu="cpu0"} 28675.47
cpu_time_system{cpu="cpu1"} 27779.34
cpu_time_system{cpu="cpu2"} 27406.18
cpu_time_system{cpu="cpu3"} 27404.97
# HELP cpu_time_user Telegraf collected metric
# TYPE cpu_time_user counter
cpu_time_user{cpu="cpu0"} 99551.84
cpu_time_user{cpu="cpu1"} 103468.52
cpu_time_user{cpu="cpu2"} 102591.45
cpu_time_user{cpu="cpu3"} 100717.05
```
