# Telegraf Output Buffer Strategy

## Objective

Introduce a new agent-level config option to choose a disk buffer strategy for
output plugin metric queues.

## Overview

Currently when a Telegraf output metric queue fills, either due to incoming
metrics being too fast or various issues with writing to the output, such as
connection failures or rate limiting, new metrics are dropped and never
written to the output. This specification defines a set of options to make
this output queue more durable by persisting pending metrics to disk rather
than only an in-memory limited size queue.

## Keywords

output plugins, agent configuration, persist to disk

## Agent Configuration

The configuration is at the agent-level, with options for:
- **Memory**, the current implementation, with no persistance to disk
- **Write-through**, all metrics are also written to disk using a
  Write Ahead Log (WAL) file
- **Disk-overflow**, when the memory buffer fills, metrics are flushed to a
  WAL file to avoid dropping overflow metrics

As well as an option to specify a directory to store the WAL files on disk,
with a default value. These configurations are global, and no change means
memory only mode, retaining current behavior.

## Metric Ordering and Tracking

Tracking metrics will be accepted either on a successful write to the output
source like currently, or on write to the WAL file in the case of the
disk-overflow option. Metrics will be written to their appropriate output in
the order they are received in the buffer still no matter which buffer
strategy is chosen.

## Disk Utilization and File Handling

Each output plugin has its own in-memory output buffer, and therefore will
each have their own WAL file (or potentially files) for buffer persistence.
Telegraf will not make any attempt to limit the size on disk taken by these
files, beyond cleaning up WAL files for metrics that have successfully been
flushed to their output source. It is the user's responsibility to ensure
these files do not entirely fill the disk, both during Telegraf uptime and
with lingering files from previous instances of the program.

Telegraf should provide a way to easily flush WAL files from previous
instances of the program in the event that a crash or system failure
happens. Telegraf makes no guarantee that in these cases, all metrics will
be kept. This may be as simple as a plugin which can read these WAL files
as an input. The file names should be clear to the user what order they are
in so that if metric order for writing to output is crucial, it can be
retained. This plugin should not be required for use to allow the buffer
strategy to work at all, but as a backup option for the user in the event
that files linger across multiple runs of Telegraf.

## Is/Is-not
- Is a way to increase the durability of metrics and reduce the potential
  for metrics to be dropped due to a full in-memory buffer
- Is not a way to guarantee data safety in the event of a crash or system failure
- Is not a way to manage file system allocation size, file space will be used
  until the disk is full

## Prior art

[Initial issue](https://github.com/influxdata/telegraf/issues/802)
[Loose specification issue](https://github.com/influxdata/telegraf/issues/14805)
