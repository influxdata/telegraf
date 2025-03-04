# Batch Processor Plugin

This processor groups metrics into batches by adding a batch tag. This is
useful for parallel processing of metrics where downstream processors,
aggregators or outputs can then select a batch using `tagpass` or `metricpass`.

Metrics are distributed across batches using the round-robin scheme.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
## Batch metrics into separate batches by adding a tag indicating the batch index.
[[processors.batch]]
  ## The name of the tag to use for adding the batch index
  batch_tag = "my_batch"

  ## The number of batches to create
  batches = 16

  ## Do not assign metrics with an existing batch assignment to a
  ## different batch. 
  # skip_existing = false
```

## Example

The example below uses these settings:

```toml
[[processors.batch]]
  ## The tag key to use for batching
  batch_tag = "batch"
  
  ## The number of batches to create
  batches = 3
```

```diff
- temperature cpu=25
- temperature cpu=50
- temperature cpu=75
- temperature cpu=25
- temperature cpu=50
- temperature cpu=75
+ temperature,batch=0 cpu=25
+ temperature,batch=1 cpu=50
+ temperature,batch=2 cpu=75
+ temperature,batch=0 cpu=25
+ temperature,batch=1 cpu=50
+ temperature,batch=2 cpu=75
```
