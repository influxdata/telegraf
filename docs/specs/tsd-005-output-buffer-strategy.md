# Telegraf Output Buffer Strategy

## Objective

Introduce a new agent-level config option to choose a disk buffer strategy for
output plugin metric queues.

## Overview

Currently, when a Telegraf output metric queue fills, either due to incoming
metrics being too fast or various issues with writing to the output, new
metrics are dropped and never written to the output. This specification
defines a set of options to make this output queue more durable by persisting
pending metrics to disk rather than only an in-memory limited size queue.

## Keywords

output plugins, agent configuration, persist to disk

## Agent Configuration

The configuration is at the agent-level, with options for:

- **Memory**, the current implementation, with no persistence to disk
- **Write-through**, all metrics are also written to disk using a
  Write Ahead Log (WAL) file
- **Disk-overflow**, when the memory buffer fills, metrics are flushed to a
  WAL file to avoid dropping overflow metrics

As well as an option to specify a directory to store the WAL files on disk,
with a default value. These configurations are global, and no change means
memory only mode, retaining current behavior.

## Metric Ordering and Tracking

Tracking metrics will be accepted either on a successful write to the output
source like currently, or on write to the WAL file. Metrics will be written
to their appropriate output in the order they are received in the buffer
regardless of which buffer strategy is chosen.

## Disk Utilization and File Handling

Each output plugin has its own in-memory output buffer, and therefore will
each have their own WAL file for buffer persistence. This file may not exist
if Telegraf is successfully able to write all of its metrics without filling
the in-memory buffer in disk-overflow mode, or not at all in memory mode.
Telegraf should use one file per output plugin, and remove entries from the
WAL file as they are written to the output.

Telegraf will not make any attempt to limit the size on disk taken by these
files beyond cleaning up WAL files for metrics that have successfully been
flushed to their output source. It is the user's responsibility to ensure
these files do not entirely fill the disk, both during Telegraf uptime and
with lingering files from previous instances of the program.

If WAL files exist for an output plugin from previous instances of Telegraf,
they will be picked up and flushed before any new metrics that are written
to the output. This is to ensure that these metrics are not lost, and to
ensure that output write order remains consistent.

Telegraf must additionally provide a way to manually flush WAL files via
some separate plugin or similar. This could be used as a way to ensure that
WAL files are properly written in the event that the output plugin changes
and the WAL file is unable to be detected by a new instance of Telegraf.
This plugin should not be required for use to allow the buffer strategy to
work.

## Is/Is-not

- Is a way to prevent metrics from being dropped due to a full memory buffer
- Is not a way to guarantee data safety in the event of a crash or system failure
- Is not a way to manage file system allocation size, file space will be used
  until the disk is full

## Prior art

[Initial issue](https://github.com/influxdata/telegraf/issues/802)
[Loose specification issue](https://github.com/influxdata/telegraf/issues/14805)
