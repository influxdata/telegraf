# Fluentd Input Plugin

The fluentd plugin gathers metrics from plugin endpoint provided by [in_monitor
plugin][1].  This plugin understands data provided by /api/plugin.json resource
(/api/config.json is not covered).

You might need to adjust your fluentd configuration, in order to reduce series
cardinality in case your fluentd restarts frequently. Every time fluentd starts,
`plugin_id` value is given a new random value.  According to [fluentd
documentation][2], you are able to add `@id` parameter for each plugin to avoid
this behaviour and define custom `plugin_id`.

example configuration with `@id` parameter for http plugin:

```text
<source>
  @type http
  @id http
  port 8888
</source>
```

[1]: https://docs.fluentd.org/input/monitor_agent
[2]: https://docs.fluentd.org/configuration/config-file#common-plugin-parameter

## Configuration

```toml @sample.conf
# Read metrics exposed by fluentd in_monitor plugin
[[inputs.fluentd]]
  ## This plugin reads information exposed by fluentd (using /api/plugins.json endpoint).
  ##
  ## Endpoint:
  ## - only one URI is allowed
  ## - https is not supported
  endpoint = "http://localhost:24220/api/plugins.json"

  ## Define which plugins have to be excluded (based on "type" field - e.g. monitor_agent)
  exclude = [
    "monitor_agent",
    "dummy",
  ]
```

## Measurements & Fields

Fields may vary depending on the plugin type

- fluentd
  - retry_count              (float, unit)
  - buffer_queue_length      (float, unit)
  - buffer_total_queued_size (float, unit)
  - rollback_count           (float, unit)
  - flush_time_count         (float, unit)
  - slow_flush_count         (float, unit)
  - emit_count               (float, unit)
  - emit_records             (float, unit)
  - emit_size                (float, unit)
  - write_count              (float, unit)
  - buffer_stage_length      (float, unit)
  - buffer_queue_byte_size   (float, unit)
  - buffer_stage_byte_size   (float, unit)
  - buffer_available_buffer_space_ratios (float, unit)  

## Tags

- All measurements have the following tags:
  - plugin_id        (unique plugin id)
  - plugin_type      (type of the plugin e.g. s3)
    - plugin_category  (plugin category e.g. output)

## Example Output

```shell
$ telegraf --config fluentd.conf --input-filter fluentd --test
* Plugin: inputs.fluentd, Collection 1
> fluentd,host=T440s,plugin_id=object:9f748c,plugin_category=input,plugin_type=dummy buffer_total_queued_size=0,buffer_queue_length=0,retry_count=0 1492006105000000000
> fluentd,plugin_category=input,plugin_type=dummy,host=T440s,plugin_id=object:8da98c buffer_queue_length=0,retry_count=0,buffer_total_queued_size=0 1492006105000000000
> fluentd,plugin_id=object:820190,plugin_category=input,plugin_type=monitor_agent,host=T440s retry_count=0,buffer_total_queued_size=0,buffer_queue_length=0 1492006105000000000
> fluentd,plugin_id=object:c5e054,plugin_category=output,plugin_type=stdout,host=T440s buffer_queue_length=0,retry_count=0,buffer_total_queued_size=0 1492006105000000000
> fluentd,plugin_type=s3,host=T440s,plugin_id=object:bd7a90,plugin_category=output buffer_queue_length=0,retry_count=0,buffer_total_queued_size=0 1492006105000000000
> fluentd,plugin_id=output_td, plugin_category=output,plugin_type=tdlog, host=T440s buffer_available_buffer_space_ratios=100,buffer_queue_byte_size=0,buffer_queue_length=0,buffer_stage_byte_size=0,buffer_stage_length=0,buffer_total_queued_size=0,emit_count=0,emit_records=0,flush_time_count=0,retry_count=0,rollback_count=0,slow_flush_count=0,write_count=0 1651474085000000000
```
