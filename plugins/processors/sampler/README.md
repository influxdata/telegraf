The Sampler processor will sample a percentage of random metrics over a given field.

The `sample_field` should be a unique identifier, as the processor will use that
to determine which metrics get passed through.

```
[[processors.sampler]]

## integer representing the percentage of metrics
## to be passed through
percent_of_metrics = 5

## field to be sampled over
sample_field = "trace_id"
```