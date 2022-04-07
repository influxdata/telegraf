# Hashicorp Consul Agent Metrics Input Plugin

This plugin grabs metrics from a Consul agent. Telegraf may be present in every node and connect to the agent locally. In this case should be something like `http://127.0.0.1:8500`.

> Tested on Consul 1.10.4 .

## Configuration

```toml
# Read metrics from the Consul Agent API
[[inputs.consul_agent]]
  ## URL for the Consul agent
  # url = "http://127.0.0.1:8500"

  ## Use auth token for authorization.
  ## If both are set, an error is thrown.
  ## If both are empty, no token will be used.
  # token_file = "/path/to/auth/token"
  ## OR
  # token = "a1234567-40c7-9048-7bae-378687048181"

  ## Set timeout (default 5 seconds)
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
```

## Metrics

Consul collects various metrics. For every details, please have a look at Consul following documentation:

- [https://www.consul.io/api/agent#view-metrics](https://www.consul.io/api/agent#view-metrics)
