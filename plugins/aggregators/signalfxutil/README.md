# SignalFx Utilization Plugin

```toml
[[aggregators.signalfx_util]]
  ## SignalFx Utilization Aggregator
  ## Enable this plugin to report utilization metrics
  ## Metrics will report with the plugin name "signalfx-metadata"
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  ## The period must be at least double the collection interval
  ## because this plugin aggregates metrics across two reporting intervals.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## Only pass the following metrics to the utilization plugin
  namepass = ["cpu", "mem", "disk", "diskio", "net", "system"]
```