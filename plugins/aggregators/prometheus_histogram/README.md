# Prometheus Histogram Aggregator Plugin

The histogram aggregator plugin creates prometheus compatible histograms containing the counts of
gauge field values within a range.

Values added to a bucket are also added to the larger buckets in the
distribution.  This creates a [cumulative histogram](https://en.wikipedia.org/wiki/Histogram#/media/File:Cumulative_vs_normal_histogram.svg).

Like other Telegraf aggregators, the metric is emitted every `period` seconds.
Bucket counts however are not reset between periods and will be non-strictly
increasing while Telegraf is running.

#### Design

Each metric is passed to the aggregator and this aggregator creates a prometheus histogram
for metrics defined in the aggregator configuration.  You can use the "namepass" property to limit the metrics
passed to the aggregator.  Each metric passed to the aggregator with a "gauge" field for the specified metric name will
be treated as a single prometheus histogram observation.

### Configuration

```toml
# Configuration for aggregate histogram metrics
[[aggregators.histogram]]
  ## The period in which to flush the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Example config that aggregates a metric.
  [[aggregators.prometheus_histogram.config]]
    ## The set of buckets.
    buckets = [0.0, 15.6, 34.5, 49.1, 71.5, 80.5, 94.5, 100.0]
    ## The name of metric.
    measurement_name = "cpu"
    ## Unit of the measurement
    measurement_unit = "seconds"
```

The user is responsible for defining the bounds of the histogram bucket as
well as the measurement name.

Each histogram config section must contain a `buckets`, `measurement_name`, and
`measurement_unit` option.

The `buckets` option contains a list of floats which specify the bucket
boundaries.  Each float value defines the inclusive upper bound of the bucket.
The `+Inf` bucket is added automatically and does not need to be defined.