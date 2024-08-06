# Output rate limiting

## Objective

Allow to control the metric-rate sent by outputs

## Keywords

output plugins, rate limit, buffer

## Overview

Output plugins send metrics to their corresponding services respecting the
`metric_batch_size` and the `flush_interval` configured. While this works well
in most situations, special situations might occur where the output will send
a large number of metrics in a short time-span. E.g. when a large number of
metrics are gathered in a short amount of time by one or more inputs or when
reconnecting after a longer disconnect of an output from it's service.
In all of those cases a large number of batches are prepared and sent via the
output plugin to its service potentially overwhelming the service in turn with
the number of metrics sent.
Furthermore, use-cases exist where operators want to provision limited resources
to Telegraf and in turn want to control the data-rate to a service.

This specification intends to introduce an _optional_ rate limiting feature
configurable per output to gain control of the sending rate of output plugins.
Therefore, a new `metric_rate_limit` setting is proposed allowing to set the
maximum number of metrics sent __per second__ via an output. By default, the
metric rate must be unlimited.

In case the specified rate limit is reached, a smaller batch satisfying the
limit is sent or, if no metrics are left, the write-cycle is skipped by the
output. The user should be informed in the logs if the rate limit applies.

## Caveats

It is important to note that setting a metric rate limit poses a severe
constraint for an output, so the feature should be used carefully. Please make
sure the configured metric rate limit exceeds the average input rate of metrics
gathered by inputs.
In case the limit is set too low, i.e. below the average rate metrics are
gathered by inputs, the output might not be able to sent the metrics fast
enough. In turn the metrics buffer will fill up and metrics are dropped.
Telegraf might not be able to recover from this situation in case the output
rate is permanently below the input rate.

## Related Issues

- [#15353](https://github.com/influxdata/telegraf/issues/15353) rate limiting processor proposal
