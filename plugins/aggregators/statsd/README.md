# Statsd Aggregator Plugin

The Statsd aggregator plugin deal with metrics from statsd.

For gauge, aggregator update with latest metric.
Value with a sing will change the value, rather than settings it.
```
gaugor:10|g
gaugor:5|g
gaugor:+1|g <-- would result in value 6
```

For counter, same metric's value is be accumulated and reset to 0 at each flush.
```
counter:1|g
counter:2|g
counter:1|g <-- would result in value 4
```

For set, aggregator count unique occurences of events between pushes.
```
user:a|s
user:b|s
user:a|s <-- would result in value 2
```

For timing & histogram, the following extra fields are made:

- `lower`: The lower bound is the lowest value statsd saw for that stat during that interval.
- `upper`: The upper bound is the highest value statsd saw for that stat during that interval.
- `mean`: The mean is the average of all values statsd saw for that stat during that interval.
- `stddev`: The stddev is the sample standard deviation of all values statsd saw for that stat during that interval.
- `sum`: The sum is the sample sum of all values statsd saw for that stat during that interval.
- `count`: The count is the number of timings statsd saw for that stat during that interval. It is not averaged.
- `percentiles_<P>`: The Pth percentile is a value x such that P% of all the values statsd saw for that stat during that time period are below x. The most common value that people use for P is the 90, this is a great number to try to optimize.

### Configuration
```toml
[[aggregators.statsd]]
  # General Aggregator Arguments:

  ## The period on which to flush & clear the aggregator.
  period = "5s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = true

  ## this aggregator only can deal with metrics from statsd
  [[aggregators.statsd.tagpass]]
    statsd_type = ["g", "c", "s", "ms", "h"]

  # Statsd Arguments:

  ## The following configuration options control when aggregator clears it's
  ## cache of previous values. If set to false, then telegraf will only clear
  ## it's cache when the daemon is restarted.
  ## Reset gauges every interval (default=true)
  delete_gauges = true
  ## Reset counters every interval (default=true)
  delete_counters = true
  ## Reset sets every interval (default=true)
  delete_sets = true
  ## Reset timings & histograms every interval (default=true)
  delete_timings = true

  ## Percentiles to calculate for timing & histogram stats
  percentiles = [90]

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000
```
