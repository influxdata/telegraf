# Groundwork Output Plugin

This plugin writes to a [GroundWork Monitor][1] instance. Plugin only supports GW8+

[1]: https://www.gwos.com/product/groundwork-monitor/

### Configuration:

```toml
[[outputs.groundwork]]
  ## HTTP endpoint for your groundwork instance.
  # groundwork_endpoint = ""

  ## Agent uuid for Groundwork API Server
  # agent_id = ""

  ## Username to access Groundwork API
  # username = ""
  
  ## Password to use in pair with username
  # password = ""
  
  ## Default display name for the host with services(metrics) [default - "telegraf"]
  # default_host = "telegraf"

  ## Default service state [default - "SERVICE_OK"]
  # default_service_state = "SERVICE_OK"

  ## The name of the tag that contains the hostname [default - "host"]
  # resource_tag = "host"
```

### List of tags used by the plugin:

```
	• service  - to define the name of the service you want to monitor
	• status   - to define the status of the service
	• message  - to provide any message you want
	• unitType - to use in monitoring contexts(subset of The Unified Code for Units of Measure standard)
	• warning  - to define warning threshold value
	• critical - to define critical threshold value
```


