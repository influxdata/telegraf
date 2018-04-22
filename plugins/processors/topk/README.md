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

  ## Over which tags should the aggregation be done. Globs can be specified, in
  ## which case any tag matching the glob will aggregated over. If set to an
  ## empty list is no aggregation over tags is done
  # group_by = ['*']

  ## The plugin can aggregate over several fields. If more than one field is
  ## specified, an aggregation is calculated per group per field.

  ## The plugin returns a metric if it's in a group in the top k groups,
  ## ordered by any of the aggregations of the selected fields

  ## This effectively means that more than K metrics may be returned. If you
  ## need to return only the top k metrics regardless of grouping, use the simple_topk setting


  ## Over which fields are the top k are calculated
  # fields = ["value"]

  ## What aggregation to use. Options: sum, mean, min, max
  # aggregation = "mean"

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
  # add_groupby_tag = ""

  ## These settings provide a way to know the position of each metric in
  ## the top k. The 'add_rank_field' setting allows to specify for which
  ## fields the position is required. If the list is non empty, then a field
  ## will be added to each every metric for each field present in the
  ## 'add_rank_field'. This field will contain the ranking of the group that
  ## the metric belonged to when aggregated over that field.
  ## The name of the field will be set to the name of the aggregation field,
  ## suffixed by the value of the 'rank_field_suffix' setting
  # add_rank_fields = []
  # rank_field_suffix = "_rank"

  ## These settings provide a way to know what values the plugin is generating
  ## when aggregating metrics. The 'add_agregate_field' setting allows to
  ## specify for which fields the final aggregation value is required. If the
  ## list is non empty, then a field will be added to each every metric for
  ## each field present in the 'add_aggregate_field'. This field will contain
  ## the computed aggregation for the group that the metric belonged to when
  ## aggregated over that field.
  ## The name of the field will be set to the name of the aggregation field,
  ## suffixed by the value of the 'aggregate_field_suffix' setting
  # add_aggregate_fields = []
  # aggregate_field_suffix = "_rank"
```

### Tags:

This processor does not add tags by default. But the setting `add_groupby_tag` will add a tag if set to anything other than ""


### Fields:

This processor does not add fields by default. But the settings `add_rank_fields` and `add_aggregation_fields` will add one or several fields if set to anything other than ""
