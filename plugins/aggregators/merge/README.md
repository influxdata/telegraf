# Merge Aggregator

Merge metrics together into a metric with multiple fields into the most memory
and network transfer efficient form.

Use this plugin when fields are split over multiple metrics, with the same
measurement, tag set and timestamp.  By merging into a single metric they can
be handled more efficiently by the output.

## Configuration

```toml
# Merge metrics into multifield metrics by series key
[[aggregators.merge]]
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = true
```

## Example

```diff
- cpu,host=localhost usage_time=42 1567562620000000000
- cpu,host=localhost idle_time=42 1567562620000000000
+ cpu,host=localhost idle_time=42,usage_time=42 1567562620000000000
```
