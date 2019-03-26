# TopK Processor Plugin

The TopK processor plugin is a filter designed to get the top series over a period of time. It can be tweaked to do its top k computation over a period of time, so spikes can be smoothed out.

This processor goes through these steps when processing a batch of metrics:

  1. Groups metrics in buckets using their tags and name as key
  2. Aggregates each of the selected fields for each bucket by the selected aggregation function (sum, mean, etc)
  3. Orders the buckets by one of the generated aggregations, returns all metrics in the top `K` buckets, then reorders the buckets by the next of the generated aggregations, returns all metrics in the top `K` buckets, etc, etc, etc, until it runs out of fields.

The plugin makes sure not to duplicate metrics

Note that depending on the amount of metrics on each computed bucket, more than `K` metrics may be returned

### Configuration:

```toml
[[processors.topk]]
  ## How many seconds between aggregations
  # period = 10

  ## How many top metrics to return
  # k = 10

  ## Over which tags should the aggregation be done. Globs can be specified, in
  ## which case any tag matching the glob will aggregated over. If set to an
  ## empty list is no aggregation over tags is done
  # group_by = ['*']

  ## Over which fields are the top k are calculated
  # fields = ["value"]

  ## What aggregation to use. Options: sum, mean, min, max
  # aggregation = "mean"

  ## Instead of the top k largest metrics, return the bottom k lowest metrics
  # bottomk = false

  ## The plugin assigns each metric a GroupBy tag generated from its name and
  ## tags. If this setting is different than "" the plugin will add a
  ## tag (which name will be the value of this setting) to each metric with
  ## the value of the calculated GroupBy tag. Useful for debugging
  # add_groupby_tag = ""

  ## These settings provide a way to know the position of each metric in
  ## the top k. The 'add_rank_field' setting allows to specify for which
  ## fields the position is required. If the list is non empty, then a field
  ## will be added to each and every metric for each string present in this
  ## setting. This field will contain the ranking of the group that
  ## the metric belonged to when aggregated over that field.
  ## The name of the field will be set to the name of the aggregation field,
  ## suffixed with the string '_topk_rank'
  # add_rank_fields = []

  ## These settings provide a way to know what values the plugin is generating
  ## when aggregating metrics. The 'add_agregate_field' setting allows to
  ## specify for which fields the final aggregation value is required. If the
  ## list is non empty, then a field will be added to each every metric for
  ## each field present in this setting. This field will contain
  ## the computed aggregation for the group that the metric belonged to when
  ## aggregated over that field.
  ## The name of the field will be set to the name of the aggregation field,
  ## suffixed with the string '_topk_aggregate'
  # add_aggregate_fields = []
```

### Tags:

This processor does not add tags by default. But the setting `add_groupby_tag` will add a tag if set to anything other than ""


### Fields:

This processor does not add fields by default. But the settings `add_rank_fields` and `add_aggregation_fields` will add one or several fields if set to anything other than ""


### Example
**Config**
```toml
[[processors.topk]]
  period = 20
  k = 3
  group_by = ["pid"]
  fields = ["cpu_usage"]
```

**Output difference with topk**
```diff
< procstat,pid=2088,process_name=Xorg cpu_usage=7.296576662282613 1546473820000000000
< procstat,pid=2780,process_name=ibus-engine-simple cpu_usage=0 1546473820000000000
< procstat,pid=2554,process_name=gsd-sound cpu_usage=0 1546473820000000000
< procstat,pid=3484,process_name=chrome cpu_usage=4.274300361942799 1546473820000000000
< procstat,pid=2467,process_name=gnome-shell-calendar-server cpu_usage=0 1546473820000000000
< procstat,pid=2525,process_name=gvfs-goa-volume-monitor cpu_usage=0 1546473820000000000
< procstat,pid=2888,process_name=gnome-terminal-server cpu_usage=1.0224991500287577 1546473820000000000
< procstat,pid=2454,process_name=ibus-x11 cpu_usage=0 1546473820000000000
< procstat,pid=2564,process_name=gsd-xsettings cpu_usage=0 1546473820000000000
< procstat,pid=12184,process_name=docker cpu_usage=0 1546473820000000000
< procstat,pid=2432,process_name=pulseaudio cpu_usage=9.892858669796528 1546473820000000000
---
> procstat,pid=2432,process_name=pulseaudio cpu_usage=11.486933087507786 1546474120000000000
> procstat,pid=2432,process_name=pulseaudio cpu_usage=10.056503212060552 1546474130000000000
> procstat,pid=23620,process_name=chrome cpu_usage=2.098690278123081 1546474120000000000
> procstat,pid=23620,process_name=chrome cpu_usage=17.52514619948493 1546474130000000000
> procstat,pid=2088,process_name=Xorg cpu_usage=1.6016732172309973 1546474120000000000
> procstat,pid=2088,process_name=Xorg cpu_usage=8.481040931533833 1546474130000000000
```
