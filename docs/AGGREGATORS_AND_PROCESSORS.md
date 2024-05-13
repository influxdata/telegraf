# Aggregator & Processor Plugins

Telegraf has the concept of aggregator and processor plugins, which sit between
inputs and outputs. These plugins allow a user to do additional processing or
aggregation to collected metrics.

```text
┌───────────┐
│           │
│    CPU    │───┐
│           │   │
└───────────┘   │
                │
┌───────────┐   │                                              ┌───────────┐
│           │   │                                              │           │
│  Memory   │───┤                                          ┌──▶│ InfluxDB  │
│           │   │                                          │   │           │
└───────────┘   │    ┌─────────────┐     ┌─────────────┐   │   └───────────┘
                │    │             │     │Aggregators  │   │
┌───────────┐   │    │Processors   │     │ - mean      │   │   ┌───────────┐
│           │   │    │ - transform │     │ - quantiles │   │   │           │
│   MySQL   │───┼───▶│ - decorate  │────▶│ - min/max   │───┼──▶│   File    │
│           │   │    │ - filter    │     │ - count     │   │   │           │
└───────────┘   │    │             │     │             │   │   └───────────┘
                │    └─────────────┘     └─────────────┘   │
┌───────────┐   │                                          │   ┌───────────┐
│           │   │                                          │   │           │
│   SNMP    │───┤                                          └──▶│   Kafka   │
│           │   │                                              │           │
└───────────┘   │                                              └───────────┘
                │
┌───────────┐   │
│           │   │
│  Docker   │───┘
│           │
└───────────┘
```

## Ordering

Processors are run first, then aggregators, then processors a second time.

Allowing processors to run again after aggregators gives users the opportunity
to run a processor on any aggregated metrics. This behavior can be a bit
surprising to new users and may cause weird behavior in metrics. For example,
if the user scales data, it could get scaled twice!

To disable this behavior set the `skip_processors_after_aggregators` agent
configuration setting to true. Another option is to use metric filtering as
described below.

## Metric Filtering

Use [metric filtering][] to control which metrics are passed through a processor
or aggregator.  If a metric is filtered out the metric bypasses the plugin and
is passed downstream to the next plugin.

[metric filtering]: CONFIGURATION.md#measurement-filtering

## Processor

Processor plugins process metrics as they pass through and immediately emit
results based on the values they process. For example, this could be printing
all metrics or adding a tag to all metrics that pass through.

See the [processors][] for a full list of processor plugins available.

[processors]: https://github.com/influxdata/telegraf/tree/master/plugins/processors

## Aggregator

Aggregator plugins, on the other hand, are a bit more complicated. Aggregators
are typically for emitting new _aggregate_ metrics, such as a running mean,
minimum, maximum, or standard deviation. For this reason, all _aggregator_
plugins are configured with a `period`. The `period` is the size of the window
of metrics that each _aggregate_ represents. In other words, the emitted
_aggregate_ metric will be the aggregated value of the past `period` seconds.

Since many users will only care about their aggregates and not every single
metric gathered, there is also a `drop_original` argument, which tells Telegraf
to only emit the aggregates and not the original metrics.

Since aggregates are created for each measurement, field, and unique tag
combination the plugin receives, you can make use of `taginclude` to group
aggregates by specific tags only.

See the [aggregators][] for a full list of aggregator plugins available.

**Note:** Aggregator plugins only aggregate metrics within their periods
(i.e. `now() - period`). Data with a timestamp earlier than `now() - period`
cannot be included.

[aggregators]: https://github.com/influxdata/telegraf/tree/master/plugins/aggregators
