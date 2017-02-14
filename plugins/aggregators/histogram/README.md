# Histogram Aggregator Plugin

#### Goal

This plugin was added for ability to build histograms.

#### Description

The histogram aggregator plugin aggregates values of specified metric\`s parameters. The metric is emitted every
`period` seconds. All you need to do is to specify borders of histogram buckets and parameters, for which you want to
aggregate histogram.

#### How it works

The each metric is passed to the aggregator and this aggregator searches histogram buckets for those parameters, which
have been specified in the config. If buckets are found, the aggregator will put +1 to appropriate bucket.
Otherwise, nothing will happen. Every `period` seconds these data will be pushed to output.

Also, the algorithm of hit counting to buckets was implemented on the base of the algorithm, which is implemented in
the Prometheus [client](https://github.com/prometheus/client_golang/blob/master/prometheus/histogram.go).

### Configuration

```toml
# Configuration for aggregate histogram metrics
[[aggregators.histogram]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## The example of config to aggregate histogram for all fields of specified metric.
  [[aggregators.histogram.config]]
  ## The set of buckets.
  buckets = [0.0, 15.6, 34.5, 49.1, 71.5, 80.5, 94.5, 100.0]
  ## The name of metric.
  metric_name = "cpu"

  ## The example of config to aggregate histogram for concrete fields of specified metric.
  [[aggregators.histogram.config]]
  ## The set of buckets.
  buckets = [0.0, 10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0]
  ## The name of metric.
  metric_name = "diskio"
  ## The concrete fields of metric.
  metric_fields = ["io_time", "read_time", "write_time"]
```

#### Explanation

The field `metric_fields` is the list of metric parameters. For example, the metric `cpu` has the following parameters:
usage_user, usage_system, usage_idle, usage_nice, usage_iowait, usage_irq, usage_softirq, usage_steal, usage_guest,
usage_guest_nice.

Note that histogram metrics will be pushed every `period` seconds. 
As you know telegraf calls aggregator `Reset()` func each `period` seconds. Histogram aggregator ignores `Reset()` and continues to count hits. 

#### Use cases

You can specify parameters using two cases:

 1. The specifying only metric name. In this case all parameters of metric will be aggregated.
 2. The specifying metric name and concrete parameters.
 
#### Some rules
 
 - The setting of each histogram must be in separate section with title `aggregators.histogram.config`.

 - The each value of bucket must be float value.
 
 - Don\`t include the border bucket `+Inf`. It will be done automatically.
 
### Measurements & Fields:

The postfix `bucket` will be added to each parameter.

- measurement1
    - field1_bucket
    - field2_bucket

### Tags:

All measurements have tag `le`. This tag has the border value of bucket. It means that the metric value is less or equal
to the value of this tag. For example, let assume that we have the metric value 10 and the following buckets:
[5, 10, 30, 70, 100]. Then the tag `le` will have the value 10, because the metrics value is passed into bucket with
right border value `10`.

### Example Output:

The following output will return to the Prometheus client.

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
