# TopK Processor Plugin

The TopK processor plugin is a filter that keeps only the top k values of a given series. It can be tweaked to do its top k computation over a period of time, so spikes can be smoothed out.

### Configuration:

```toml
[[processors.topk]]
  period = 10                  # How many seconds between aggregations. Default: 10
  k = 10                       # How many top metrics to return. Default: 10
  metric = "mymetric"          # Which metrics to consume. Supports regular expressions. No default. Mandatory
  tags = {"tag_name"="tag_value"}  # Map of tags regexes to match against. Default: "{}" (match all)
  fields = ["memory_rss"]      # Over which fields are the top k are calculated. Default: "value"
  aggregation = "avg"          # What aggregation to use. Default: "avg". Options: sum, avg, min, max
  group_by = ["process_name"]  # Over which tags should the aggregation be done. Default: []
  group_by_metric_name = false # Wheter or not to also group by metric name

  bottomk = false              # Instead of the top k largest metrics, return the bottom k lowest metrics
  revert_metric_match = false  # Whether or not to invert the metric name match
  revert_tag_match = false     # Whether or not to invert the tag match
```

### Tags:

No tags are applied by this processor.
