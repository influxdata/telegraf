# GroundWork Output Plugin

This plugin writes to a [GroundWork Monitor][1] instance. Plugin only supports GW8+

[1]: https://www.gwos.com/product/groundwork-monitor/

## Configuration

```toml
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
```

## List of tags used by the plugin

* service  - to define the name of the service you want to monitor.
* status   - to define the status of the service.
* message  - to provide any message you want.
* unitType - to use in monitoring contexts(subset of The Unified Code for Units of Measure standard). Supported types: "1", "%cpu", "KB", "GB", "MB".
* warning  - to define warning threshold value.
* critical - to define critical threshold value.

## NOTE

The current version of Groundworks does not support metrics whose values are strings. Such metrics will be skipped and will not be added to the final payload. You can find more context in this pull request: [#10255]( https://github.com/influxdata/telegraf/pull/10255)
