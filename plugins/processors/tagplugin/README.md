# Tagplugin

This plugin makes it possible to add a new tag to metrics based on the value of an existing tag.

This can be useful for example in cases where nodes running telegraf are configured differently, and their cpus 
or network interfaces have different roles on each individual node.

Tagging them at this level can thus make it easier to group them by category or role later.


### Configuration:

```toml
# # Tag cpu metrics based on the cpu number.
[[processors.tagplugin]]
  ## Only metrics with a name in this list will be processed by this plugin.
  ## If undefined all metrics will be processed.
  namepass = ["net"]

  ## The reference tag is the existing metric tag that is used to determine the value of the new tag.
  ## A tag will not be added to the metric if reference_tag_name is missing or empty.
  reference_tag_name = "interface"

  ## Name of the new tag given to the metric.
  ## A tag will not be added to the metric if new_tag_name is missing or empty.
  new_tag_name = "category"

  ## If the metric's value of reference_tag_name is not present in the map below, 
  ## the metric will be tagged with the default_tag value.
  ## However, the metric will not receive a tag if the default_tag is missing or empty.
  new_tag_default_value = "other"

  ## The keys in this map are the values to use for the new tag, when the reference tag value matches any
  ## of the elements in the corresponding list.
  ## All keys should be strings, and all values should be lists of strings.
  ## If this map is empty or not defined, all metrics passed through this plugin will be tagged with the 
  ## default_tag instead.
  ## Do not repeat values in different lists, if this happens a random of the matching keys will be used.
  ## Due to the way TOML is parsed, this map must be defined at the end of the plugin definition, 
  ## otherwise subsequent plugin config options will be interpreted as part of this map.
  [processors.tagplugin.new_tag_value_map]
    management = ["en0", "en1"]
    api = ["en2"]
```

It is possible to run multiple instances of this plugin with multiple metrics and independent tag value maps.
To do so, just add additional versions of the config above to your telegraf config file.