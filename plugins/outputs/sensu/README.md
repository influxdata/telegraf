# Sensu Go Output Plugin

This plugin writes metrics events to [Sensu Go](https://sensu.io) via its
HTTP events API.

## Configuration

```toml
# Send aggregate metrics to Sensu Monitor
[[outputs.sensu]]
  ## BACKEND API URL is the Sensu Backend API root URL to send metrics to
  ## (protocol, host, and port only). The output plugin will automatically
  ## append the corresponding backend API path
  ## /api/core/v2/namespaces/:entity_namespace/events/:entity_name/:check_name).
  ##
  ## Backend Events API reference:
  ## https://docs.sensu.io/sensu-go/latest/api/events/
  ##
  ## AGENT API URL is the Sensu Agent API root URL to send metrics to
  ## (protocol, host, and port only). The output plugin will automatically
  ## append the correspeonding agent API path (/events).
  ##
  ## Agent API Events API reference:
  ## https://docs.sensu.io/sensu-go/latest/api/events/
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

  ## Timeout for HTTP message
  # timeout = "5s"

  ## HTTP Content-Encoding for write request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "identity"

  ## Sensu Event details
  ##
  ## Below are the event details to be sent to Sensu.  The main portions of the
  ## event are the check, entity, and metrics specifications. For more information
  ## on Sensu events and its components, please visit:
  ## - Events - https://docs.sensu.io/sensu-go/latest/reference/events
  ## - Checks -  https://docs.sensu.io/sensu-go/latest/reference/checks
  ## - Entities - https://docs.sensu.io/sensu-go/latest/reference/entities
  ## - Metrics - https://docs.sensu.io/sensu-go/latest/reference/events#metrics
  ##
  ## Check specification
  ## The check name is the name to give the Sensu check associated with the event
  ## created. This maps to check.metatadata.name in the event.
  [outputs.sensu.check]
    name = "telegraf"

  ## Entity specification
  ## Configure the entity name and namespace, if necessary. This will be part of
  ## the entity.metadata in the event.
  ##
  ## NOTE: if the output plugin is configured to send events to a
  ## backend_api_url and entity_name is not set, the value returned by
  ## os.Hostname() will be used; if the output plugin is configured to send
  ## events to an agent_api_url, entity_name and entity_namespace are not used.
  # [outputs.sensu.entity]
  #   name = "server-01"
  #   namespace = "default"

  ## Metrics specification
  ## Configure the tags for the metrics that are sent as part of the Sensu event
  # [outputs.sensu.tags]
  #   source = "telegraf"

  ## Configure the handler(s) for processing the provided metrics
  # [outputs.sensu.metrics]
  #   handlers = ["influxdb","elasticsearch"]
```
