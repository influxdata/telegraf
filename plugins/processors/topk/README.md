# TopK Processor Plugin

The TopK processor plugin is a filter designed to get the top series over a period of time. It can be tweaked to do its top k computation over a period of time, so spikes can be smoothed out.

This plugin groups the metrics based on their name and tags, then generates aggregates values across each group base on fields selected by the user. It then sorts these groups based on these aggregations and returns any metric that belongs to a group in the top k (sorted by any of the aggregations). This means that when calculating the top k, more than k metrics may be returned.

If only the very top k metrics are needed, regardless of grouping, the simple_topk setting will force each metric into its own individual group

### Configuration:

```toml
[[processors.topk]]
  ## How many seconds between aggregations
  # period = 10

  ## How many top metrics to return
  # k = 10

  ## Metrics are grouped based on their tags and name. The plugin aggregates
  ## the selected fields of these groups of metrics and sorts the groups based
  ## these aggregations

  ## Over which tags should the aggregation be done. If set to n empty list is
  ## no aggregation over tags is done
  # group_by = ['*']

  ## Wheter or not to also group by metric name
  # group_by_metric_name = false

  ## The plugin can aggregate over several fields. If more than one field is
  ## specified, an aggregation is calculated per group per field.

  ## The plugin returns a metric if it's in a group in the top k groups,
  ## ordered by any of the aggregations of the selected fields

  ## This effectively means that more than K metrics may be returned. If you
  ## need to return only the top k metrics regardless of grouping, use the simple_topk setting


  ## Over which fields are the top k are calculated
  # fields = ["value"]

  ## What aggregation to use. Options: sum, avg, min, max
  # aggregation = "avg"

  ## Instead of the top k largest metrics, return the bottom k lowest metrics
  # bottomk = false

  ## If true, this will override any GroupBy options and assign each metric
  ## its own individual group. Default: false
  # simple_topk = false

  ## Drop any metrics that do fit in any group (due to nonexistent tags)
  # drop_no_group = true

  ## Drop the metrics that do not make the cut for the top k
  # drop_non_top = true

  ## The plugin assigns each metric a GroupBy tag generated from its name and
  ## tags. If this setting is different than "" the plugin will add a
  ## tag (which name will be the value of this setting) to each metric with
  ## the value of the calculated GroupBy tag. Useful for debugging
  # group_by_tag = ""

  ## This settings provides a way to know the position of each metric in
  ## the top k. If set to a value different than "", then a field (which name
  ## will be prefixed with the value of this setting) will be added to each
  ## every metric for each field over which an aggregation was made. This
  ## field will contain the ranking of the group that the metric
  ## belonged to. When aggregating over several fields, several fields will
  ## be added (one for each field over which the aggregation was calculated)
  # position_field = ""

  ## This setting provies a way know the what values the plugin is generating
  ## when aggregating the fields. If set to a value different than "", then a
  ## field (which name will be prefixed with the value of this setting) will
  ## be added to each metric which was part of a field aggregation. The value
  ## of the added field will be the value of the result of the aggregation
  ## operation for that metric's group. When aggregating over several fields,
  ## several fields will be added (one for each field over which the
  ## aggregation was calculated).
  # aggregation_field = ""
```

### Tags:

This processor does not add tags by default. But the setting `group_by_tag` will add a tag if set to anything other than ""


### Fields:

This processor does not add fields by default. But the settings `position_field` and `aggregation_field` will add one or several fields if set to anything other than ""
