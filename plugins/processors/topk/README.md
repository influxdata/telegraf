# TopK Processor Plugin

The TopK processor plugin is a filter designed to get the top series over a period of time. It can be tweaked to calculate the top metrics via different aggregation functions.

This processor goes through these steps when processing a batch of metrics:

  1. Groups measurements in buckets based on their tags and name
  2. Every N seconds, for each bucket, for each selected field: aggregate all the measurements using a given aggregation function (min, sum, mean, etc) and the field.
  3. For each computed aggregation: order the buckets by the aggregation, then returns all measurements in the top `K` buckets

Notes:
  * The deduplicates metrics
  * The name of the measurement is always used when grouping it
  * Depending on the amount of metrics on each  bucket, more than `K` series may be returned
  * If a measurement does not have one of the selected fields, it is dropped from the aggregation

### Configuration:

```toml
[[processors.topk]]
  ## How many seconds between aggregations
  # period = 10

  ## How many top buckets to return
  # k = 10

  ## Based on which tags should the buckets be computed. Globs can be specified.
  ## If set to an empty list tags are not considered when creating the buckets
  # group_by = ['*']

  ## Over which fields is the aggregation done
  # fields = ["value"]

  ## What aggregation function to use. Options: sum, mean, min, max
  # aggregation = "mean"

  ## Instead of the top k buckets, return the bottom k buckets
  # bottomk = false

  ## This setting provides a way to know wich metrics where group together.
  ## Add a tag (which name will be the value of this setting) to each metric.
  ## The value will be the tags used to pick its bucket.
  # add_groupby_tag = ""

  ## This setting provides a way to know the position of each metric's bucket in the top k
  ## If the list is non empty, a field will be added to each and every metric
  ## for each string present in this setting. This field will contain the ranking
  ## of the bucket that the metric belonged to when aggregated over that field.
  ## The name of the field will be set to the name of the aggregation field,
  ## suffixed with the string '_topk_rank'
  # add_rank_fields = []

  ## These settings provide a way to know what values the plugin is generating
  ## when aggregating metrics. If the list is non empty, then a field will be
  ## added to each every metric for each field present in this setting.
  ## This field will contain the computed aggregation for the bucket that the
  ## metric belonged to when aggregated over that field.
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
