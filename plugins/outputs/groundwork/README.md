# GroundWork Output Plugin

This plugin writes to a [GroundWork Monitor][1] instance. Plugin only supports
GW8+

[1]: https://www.gwos.com/product/groundwork-monitor/

## Configuration

```toml
# Send telegraf metrics to GroundWork Monitor
[[outputs.groundwork]]
  ## URL of your groundwork instance.
  url = "https://groundwork.example.com"

  ## Agent uuid for GroundWork API Server.
  agent_id = ""

  ## Username and password to access GroundWork API.
  username = ""
  password = ""

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

* group - to define the name of the group you want to monitor, can be changed
  with config.
* host - to define the name of the host you want to monitor, can be changed with
  config.
* service - to define the name of the service you want to monitor.
* status - to define the status of the service. Supported statuses:
  "SERVICE_OK", "SERVICE_WARNING", "SERVICE_UNSCHEDULED_CRITICAL",
  "SERVICE_PENDING", "SERVICE_SCHEDULED_CRITICAL", "SERVICE_UNKNOWN".
* message - to provide any message you want.
* unitType - to use in monitoring contexts(subset of The Unified Code for Units
  of Measure standard). Supported types: "1", "%cpu", "KB", "GB", "MB".
* warning - to define warning threshold value.
* critical - to define critical threshold value.

## NOTE

The current version of GroundWork Monitor does not support metrics whose values
are strings. Such metrics will be skipped and will not be added to the final
payload. You can find more context in this pull request: [#10255][].

[#10255]: https://github.com/influxdata/telegraf/pull/10255
