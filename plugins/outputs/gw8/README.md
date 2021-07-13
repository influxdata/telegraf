# Groundwork Output Plugin

This plugin writes to a Groundwork instance using "[GROUNDWORK][]" own format.

[GROUNDWORK]: https://github.com/gwos

### Configuration:

```toml
[[outputs.gw8]]
  ## HTTP endpoint for your groundwork instance.
  groundwork_endpoint = ""

  ## Agent uuid for Groundwork API Server
  agent_id = ""

  ## Groundwork application type
  app_type = ""

  ## Username to access Groundwork API
  username = ""
  ## Password to use in pair with username
  password = ""
```