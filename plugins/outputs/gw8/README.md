# Groundwork Output Plugin

This plugin writes to a Groundwork instance using "[GROUNDWORK][]" own format.

[GROUNDWORK]: https://www.gwos.com

### Configuration:

```toml
[[outputs.gw8]]
  ## HTTP endpoint for your groundwork instance.
  endpoint = ""

  ## Agent uuid for Groundwork API Server
  agent_id = ""

  ## Groundwork application type
  app_type = ""

  ## Username to access Groundwork API
  username = ""
  
  ## Password to use in pair with username
  password = ""
  
  ## Default display name for the host with services(metrics)
  default_host = "default_telegraf"

  ## Default service state [default - "SERVICE_OK"]
  default_service_state = "SERVICE_OK"

  ## The name of the tag that contains the hostname [default - "host"]
  resource_tag = "host"
```