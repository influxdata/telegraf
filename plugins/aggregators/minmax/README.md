# MinMax Aggregator Plugin

The minmax aggregator plugin aggregates min & max values of each field it sees,
emitting the aggrate every `period` seconds.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Keep the aggregate min/max of each metric passing through.
[[aggregators.minmax]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
```

## Measurements & Fields

- measurement1
  - field1_max
  - field1_min

## Tags

No tags are applied by this aggregator.

## Example Output

```text
system,host=tars load1=1.72 1475583980000000000
system,host=tars load1=1.6 1475583990000000000
system,host=tars load1=1.66 1475584000000000000
system,host=tars load1=1.63 1475584010000000000
system,host=tars load1_max=1.72,load1_min=1.6 1475584010000000000
system,host=tars load1=1.46 1475584020000000000
system,host=tars load1=1.39 1475584030000000000
system,host=tars load1=1.41 1475584040000000000
system,host=tars load1_max=1.46,load1_min=1.39 1475584040000000000
```
