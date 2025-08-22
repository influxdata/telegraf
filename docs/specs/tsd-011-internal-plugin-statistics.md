# Internal plugin statistics collection

## Objective

Provide a way to report plugin-internal statistics for collection by the
[`internal` input plugin][internal].

[internal]: /plugins/inputs/internal/README.md

## Keywords

plugins, statistics

## Overview

The [`internal` input plugin][internal] allows to collect statistics about
active plugins in Telegraf. These statistics are valuable indicators for
operating and optimizing Telegraf setups and for detecting issues.

Statistics are provided by plugin models and are then gathered by the
[`internal` input plugin][internal] to be emitted through the normal plugin
pipeline. However, not all important statistics are known on the model level.
Some are only known within the plugin instance itself such as the bytes written
to an output or certain error types. Emitting those statistics through direct
registration of a statistics object will not pick-up tags defined using the
global tags settings or aliases which are only known at the model level.

To overcome the mentioned limitation this specification defines a framework
allowing to inject a _statistics collector_ object into the plugin allowing to
register statistics including the `alias` and tags definitions.

To provide plugin-internal statistics a plugin should export a `Statistics`
member in the plugin structure. This member must be a pointer type
`*selfstat.Collector`. It is strongly recommended to define a `"-"` TOML tag
in order to avoid collisions with setting the member through user configuration.

The Telegraf model code then must inject a `selfstat.Collector` instance into
the plugin's `Statistics` member _after_ instantiating the plugin but _before_
calling the plugin's `Init` function. The injected collector instance must add
all relevant model-level information such as an optional `alias` setting or
`tags` settings.
The plugin must use the collector as proxy to register, unregister, reset and
access statistics.

## Related Issues

- [issue #4889](https://github.com/influxdata/telegraf/issues/4889) for
  emitting internal statistics for the InfluxDB output plugin
- [issue #6965](https://github.com/influxdata/telegraf/issues/6965) for
  emitting internal statistics of the Kafka output plugin
- [issue #17275](https://github.com/influxdata/telegraf/issues/17275) for
  emitting internal statistics for the InfluxDB v2 output plugin
