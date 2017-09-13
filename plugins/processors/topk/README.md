# TopK Processor Plugin

The TopK processor plugin is a filter that keeps only the top k values of a given series. It can be tweaked to do its top k computation over a period of time, so spikes can be smoothed out.

### Configuration:

```toml
# Print all metrics that pass through this filter.
[[processors.topk]]
  period=<seconds>
```

### Tags:

No tags are applied by this processor.
