# HTTP Output Plugin

This plugin is based off of the HTTP output plugin to send metrics
to the Sensu Events API.

### Configuration:

```toml
	## Configure check configurations
	[outputs.sensu-go.check]
		name = "telegraf"

	## BACKEND API URL is the Sensu Backend API root URL to send metrics to 
	## (protocol, host, and port only). The output plugin will automatically 
	## append the corresponding backend or agent API path (e.g. /events or 
	## /api/core/v2/namespaces/:entity_namespace/events/:entity_name/:check_name).
	## 
	## NOTE: if backend_api_url and agent_api_url and api_key are set, the output 
	## plugin will use backend_api_url. If backend_api_url and agent_api_url are 
	## not provided, the output plugin will default to use an agent_api_url of 
	## http://127.0.0.1:3031
	## 
  # backend_api_url = "http://127.0.0.1:8080"
	# agent_api_url = "http://127.0.0.1:3031"

	## API KEY is the Sensu Backend API token 
  ## Generate a new API token via: 
  ## 
  ## $ sensuctl cluster-role create telegraf --verb create --resource events,entities
  ## $ sensuctl cluster-role-binding create telegraf --cluster-role telegraf --group telegraf
  ## $ sensuctl user create telegraf --group telegraf --password REDACTED 
  ## $ sensuctl api-key grant telegraf
  ##
  ## For more information on Sensu RBAC profiles & API tokens, please visit: 
  ## - https://docs.sensu.io/sensu-go/latest/reference/rbac/
  ## - https://docs.sensu.io/sensu-go/latest/reference/apikeys/ 
  ## 
	# api_key = "${SENSU_API_KEY}"
	
  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
	# insecure_skip_verify = false

	# Optional
	## Set whether publishing to agent or directly to backend.
	# [[outputs.sensu-go]]
	#   api_key = "$SENSU_API_KEY"
	#   url = "http://127.0.0.1:3031"
	#
	#   [outputs.sensu.check]
	#     metadata = { name: "telegraf" }
	#     
	#   [outputs.sensu.entity]
	#     metadata = { name: "server-01", namespace: "default" }
	#
	#   [outputs.sensu.metrics]
	#     handlers = ["elasticsearch", "timescaledb"]
	#
	## Timeout for HTTP writes.
	# timeout = "5s"
```
