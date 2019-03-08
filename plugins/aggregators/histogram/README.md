# Histogram Aggregator Plugin

The histogram aggregator plugin creates histograms containing the counts of
field values within a range.

Values added to a bucket are also added to the larger buckets in the
distribution.  This creates a [cumulative histogram](https://en.wikipedia.org/wiki/Histogram#/media/File:Cumulative_vs_normal_histogram.svg).

Like other Telegraf aggregators, the metric is emitted every `period` seconds.
Bucket counts however are not reset between periods and will be non-strictly
increasing while Telegraf is running.

#### Design

Each metric is passed to the aggregator and this aggregator searches
histogram buckets for those fields, which have been specified in the
config. If buckets are found, the aggregator will increment +1 to the appropriate
bucket otherwise it will be added to the `+Inf` bucket.  Every `period`
seconds this data will be forwarded to the outputs.

The algorithm of hit counting to buckets was implemented on the base
of the algorithm which is implemented in the Prometheus
[client](https://github.com/prometheus/client_golang/blob/master/prometheus/histogram.go).

### Configuration

```toml
# Configuration for aggregate histogram metrics
[[aggregators.histogram]]
  ## The period in which to flush the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Example config that aggregates all fields of the metric.
  # [[aggregators.histogram.config]]
  #   ## The set of buckets.
  #   buckets = [0.0, 15.6, 34.5, 49.1, 71.5, 80.5, 94.5, 100.0]
  #   ## The name of metric.
  #   measurement_name = "cpu"

  ## Example config that aggregates only specific fields of the metric.
  # [[aggregators.histogram.config]]
  #   ## The set of buckets.
  #   buckets = [0.0, 10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0]
  #   ## The name of metric.
  #   measurement_name = "diskio"
  #   ## The concrete fields of metric
  #   fields = ["io_time", "read_time", "write_time"]
```

The user is responsible for defining the bounds of the histogram bucket as
well as the measurement name and fields to aggregate.

Each histogram config section must contain a `buckets` and `measurement_name`
option.  Optionally, if `fields` is set only the fields listed will be
aggregated.  If `fields` is not set all fields are aggregated.

The `buckets` option contains a list of floats which specify the bucket
boundaries.  Each float value defines the inclusive upper bound of the bucket.
The `+Inf` bucket is added automatically and does not need to be defined.

### Measurements & Fields:

The postfix `bucket` will be added to each field key.

- measurement1
    - field1_bucket
    - field2_bucket

### Tags:

All measurements are given the tag `le`. This tag has the border value of
bucket. It means that the metric value is less than or equal to the value of
this tag.  For example, let assume that we have the metric value 10 and the
following buckets: [5, 10, 30, 70, 100]. Then the tag `le` will have the value
10, because the metrics value is passed into bucket with right border value
`10`.

### Example Output:

```
cpu,cpu=cpu1,host=localhost,le=0.0 usage_idle_bucket=0i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=10.0 usage_idle_bucket=0i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=20.0 usage_idle_bucket=1i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=30.0 usage_idle_bucket=2i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=40.0 usage_idle_bucket=2i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=50.0 usage_idle_bucket=2i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=60.0 usage_idle_bucket=2i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=70.0 usage_idle_bucket=2i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=80.0 usage_idle_bucket=2i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=90.0 usage_idle_bucket=2i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=100.0 usage_idle_bucket=2i 1486998330000000000
cpu,cpu=cpu1,host=localhost,le=+Inf usage_idle_bucket=2i 1486998330000000000
```
