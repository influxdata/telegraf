# Fluentd Input Plugin

The fluentd plugin gathers metrics from plugin endpoint provided by [in_monitor plugin](http://docs.fluentd.org/v0.12/articles/monitoring).
This plugin understands data provided by /api/plugin.json resource (/api/config.json is not covered).

### Configuration:

```toml
# Read metrics exposed by fluentd in_monitor plugin
[[inputs.fluentd]]
    ##
    ## This plugin only reads information exposed by fluentd using /api/plugins.json.
    ## Tested using 'fluentd' version '0.14.9'
    ##
    ## Endpoint:
    ## - only one URI is allowed
    ## - https is not supported
    # Endpoint = "http://localhost:24220/api/plugins.json"

    ## Define which plugins have to be excluded (based on "type" field - e.g. monitor_agent)
    # exclude = [
    #   "monitor_agent",
    #   "dummy",
    # ]
```

### Measurements & Fields:

Fields may vary depends on type of the plugin

- fluentd
    - RetryCount            (float, unit)
    - BufferQueueLength     (float, unit)
    - BufferTotalQueuedSize (float, unit)

### Tags:

- All measurements have the following tags:
	- PluginID        (unique plugin id)
	- PluginType      (type of the plugin e.g. s3)
    - PluginCategory  (plugin category e.g. output)

### Example Output:

```
$ telegraf --config fluentd.conf --input-filter fluentd --test
* Plugin: inputs.fluentd, Collection 1
> fluentd,host=T440s,PluginID=object:9f748c,PluginCategory=input,PluginType=dummy BufferTotalQueuedSize=0,BufferQueueLength=0,RetryCount=0 1492006105000000000
> fluentd,PluginCategory=input,PluginType=dummy,host=T440s,PluginID=object:8da98c BufferQueueLength=0,RetryCount=0,BufferTotalQueuedSize=0 1492006105000000000
> fluentd,PluginID=object:820190,PluginCategory=input,PluginType=monitor_agent,host=T440s RetryCount=0,BufferTotalQueuedSize=0,BufferQueueLength=0 1492006105000000000
> fluentd,PluginID=object:c5e054,PluginCategory=output,PluginType=stdout,host=T440s BufferQueueLength=0,RetryCount=0,BufferTotalQueuedSize=0 1492006105000000000
> fluentd,PluginType=s3,host=T440s,PluginID=object:bd7a90,PluginCategory=output BufferQueueLength=0,RetryCount=0,BufferTotalQueuedSize=0 1492006105000000000

```