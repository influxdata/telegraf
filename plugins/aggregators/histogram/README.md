# Histogram Aggregator Plugin

This plugin creates histograms containing the counts of field values within the
configured range. The histogram metric is emitted every `period`.

In `cumulative` mode, values added to a bucket are also added to the
consecutive buckets in the distribution creating a [cumulative histogram][1].

> [!NOTE]
> By default bucket counts are not reset between periods and will be
> non-strictly increasing while Telegraf is running. This behavior can be
> by setting the `reset` parameter.

⭐ Telegraf v1.4.0
🏷️ statistics
💻 all

[1]: https://en.wikipedia.org/wiki/Histogram#/media/File:Cumulative_vs_normal_histogram.svg

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Configuration for aggregate histogram metrics
[[aggregators.histogram]]
  ## The period in which to flush the aggregator.
  # period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  # drop_original = false

  ## If true, the histogram will be reset on flush instead
  ## of accumulating the results.
  reset = false

  ## Whether bucket values should be accumulated. If set to false, "gt" tag will be added.
  ## Defaults to true.
  cumulative = true

  ## Expiration interval for each histogram. The histogram will be expired if
  ## there are no changes in any buckets for this time interval. 0 == no expiration.
  # expiration_interval = "0m"

  ## If true, aggregated histogram are pushed to output only if it was updated since
  ## previous push. Defaults to false.
  # push_only_on_update = false

  ## Example config that aggregates all fields of the metric.
  # [[aggregators.histogram.config]]
  #   ## Right borders of buckets (with +Inf implicitly added).
  #   buckets = [0.0, 15.6, 34.5, 49.1, 71.5, 80.5, 94.5, 100.0]
  #   ## The name of metric.
  #   measurement_name = "cpu"

  ## Example config that aggregates only specific fields of the metric.
  # [[aggregators.histogram.config]]
  #   ## Right borders of buckets (with +Inf implicitly added).
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
boundaries.  Each float value defines the inclusive upper (right) bound of the
bucket.  The `+Inf` bucket is added automatically and does not need to be
defined.  (For left boundaries, these specified bucket borders and `-Inf` will
be used).

## Measurements & Fields

The postfix `bucket` will be added to each field key.

- measurement1
  - field1_bucket
  - field2_bucket

### Tags

- `cumulative = true` (default):
  - `le`: Right bucket border. It means that the metric value is less than or
    equal to the value of this tag. If a metric value is sorted into a bucket,
    it is also sorted into all larger buckets. As a result, the value of
    `<field>_bucket` is rising with rising `le` value. When `le` is `+Inf`,
    the bucket value is the count of all metrics, because all metric values are
    less than or equal to positive infinity.
- `cumulative = false`:
  - `gt`: Left bucket border. It means that the metric value is greater than
    (and not equal to) the value of this tag.
  - `le`: Right bucket border. It means that the metric value is less than or
    equal to the value of this tag.
  - As both `gt` and `le` are present, each metric is sorted in only exactly
    one bucket.

## Example Output

Let assume we have the buckets [0, 10, 50, 100] and the following field values
for `usage_idle`: [50, 7, 99, 12]

With `cumulative = true`:

```text
cpu,cpu=cpu1,host=localhost,le=0.0 usage_idle_bucket=0i 1486998330000000000  # none
cpu,cpu=cpu1,host=localhost,le=10.0 usage_idle_bucket=1i 1486998330000000000  # 7
cpu,cpu=cpu1,host=localhost,le=50.0 usage_idle_bucket=2i 1486998330000000000  # 7, 12
cpu,cpu=cpu1,host=localhost,le=100.0 usage_idle_bucket=4i 1486998330000000000  # 7, 12, 50, 99
cpu,cpu=cpu1,host=localhost,le=+Inf usage_idle_bucket=4i 1486998330000000000  # 7, 12, 50, 99
```

With `cumulative = false`:

```text
cpu,cpu=cpu1,host=localhost,gt=-Inf,le=0.0 usage_idle_bucket=0i 1486998330000000000  # none
cpu,cpu=cpu1,host=localhost,gt=0.0,le=10.0 usage_idle_bucket=1i 1486998330000000000  # 7
cpu,cpu=cpu1,host=localhost,gt=10.0,le=50.0 usage_idle_bucket=1i 1486998330000000000  # 12
cpu,cpu=cpu1,host=localhost,gt=50.0,le=100.0 usage_idle_bucket=2i 1486998330000000000  # 50, 99
cpu,cpu=cpu1,host=localhost,gt=100.0,le=+Inf usage_idle_bucket=0i 1486998330000000000  # none
```
