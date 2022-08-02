# GroundWork Output Plugin

This plugin writes to a [GroundWork Monitor][1] instance. Plugin only supports
GW8+

[1]: https://www.gwos.com/product/groundwork-monitor/

## Configuration

```toml @sample.conf
# Send telegraf metrics to GroundWork Monitor
[[outputs.groundwork]]
  ## URL of your groundwork instance.
  url = "https://groundwork.example.com"

  ## Agent uuid for GroundWork API Server.
  agent_id = ""

  ## Username and password to access GroundWork API.
  username = ""
  password = ""

  ## Default application type to use in GroundWork client
  # default_app_type = "TELEGRAF"

  ## Default display name for the host with services(metrics).
  # default_host = "telegraf"

  ## Default service state.
  # default_service_state = "SERVICE_OK"

  ## The name of the tag that contains the hostname.
  # resource_tag = "host"

  ## The name of the tag that contains the host group name.
  # group_tag = "group"
```

## List of tags used by the plugin

* __group__ - to define the name of the group you want to monitor,
  can be changed with config.
* __host__ - to define the name of the host you want to monitor,
  can be changed with config.
* __service__ - to define the name of the service you want to monitor.
* __status__ - to define the status of the service. Supported statuses:
  "SERVICE_OK", "SERVICE_WARNING", "SERVICE_UNSCHEDULED_CRITICAL",
  "SERVICE_PENDING", "SERVICE_SCHEDULED_CRITICAL", "SERVICE_UNKNOWN".
* __message__ - to provide any message you want,
  it overrides __message__ field value.
* __unitType__ - to use in monitoring contexts (subset of The Unified Code for
  Units of Measure standard). Supported types: "1", "%cpu", "KB", "GB", "MB".
* __critical__ - to define the default critical threshold value,
  it overrides value_cr field value.
* __warning__ - to define the default warning threshold value,
  it overrides value_wn field value.
* __value_cr__ - to define critical threshold value,
  it overrides __critical__ tag value and __value_cr__ field value.
* __value_wn__ - to define warning threshold value,
  it overrides __warning__ tag value and __value_wn__ field value.

## NOTE

The current version of GroundWork Monitor does not support metrics whose values
are strings. Such metrics will be skipped and will not be added to the final
payload. You can find more context in this pull request: [#10255][].

[#10255]: https://github.com/influxdata/telegraf/pull/10255
