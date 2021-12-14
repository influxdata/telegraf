# Telegraf Aggregator & Processor Plugins

As of release 1.1.0, Telegraf has the concept of Aggregator and Processor Plugins.

These plugins sit in-between Input & Output plugins, aggregating and processing
metrics as they pass through Telegraf:

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

Both Aggregators and Processors analyze metrics as they pass through Telegraf.

Use [measurement filtering](CONFIGURATION.md#measurement-filtering)
to control which metrics are passed through a processor or aggregator.  If a
metric is filtered out the metric bypasses the plugin and is passed downstream
to the next plugin.

## Processor

Processor plugins process metrics as they pass through and immediately emit
results based on the values they process. For example, this could be printing
all metrics or adding a tag to all metrics that pass through.

## Aggregator

Aggregator plugins, on the other hand, are a bit more complicated. Aggregators
are typically for emitting new _aggregate_ metrics, such as a running mean,
minimum, maximum, or standard deviation. For this reason, all _aggregator_
plugins are configured with a `period`. The `period` is the size of the window
of metrics that each _aggregate_ represents. In other words, the emitted
_aggregate_ metric will be the aggregated value of the past `period` seconds.
Since many users will only care about their aggregates and not every single metric
gathered, there is also a `drop_original` argument, which tells Telegraf to only
emit the aggregates and not the original metrics.

Since aggregates are created for each measurement, field, and unique tag combination
the plugin receives, you can make use of `taginclude` to group
aggregates by specific tags only.

**Note:** Aggregator plugins only aggregate metrics within their periods (`now() - period`). Data with a timestamp earlier than `now() - period` cannot be included.
