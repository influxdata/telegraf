# Hashicorp Vault Input Plugin

The Vault plugin could grab metrics from every Vault agent of the cluster. Telegraf may be present in every node and connect to the agent locally. In this case should be something like `http://127.0.0.1:8200`.

> Tested on vault 1.8.5

## Configuration

```toml
# Read metrics from the Vault API
[[inputs.vault]]
  ## URL for the Vault agent
  # url = "http://127.0.0.1:8200"

  ## Use Vault token for authorization.
  ## Vault token configuration is mandatory.
  ## If both are empty or both are set, an error is thrown.
  # token_file = "/path/to/auth/token"
  ## OR
  token = "s.CDDrgg5zPv5ssI0Z2P4qxJj2"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
```

## Metrics

For a more deep understanding of Vault monitoring, please have a look at the following Vault documentation:

- [https://www.vaultproject.io/docs/internals/telemetry](https://www.vaultproject.io/docs/internals/telemetry)
- [https://learn.hashicorp.com/tutorials/vault/monitor-telemetry-audit-splunk?in=vault/monitoring](https://learn.hashicorp.com/tutorials/vault/monitor-telemetry-audit-splunk?in=vault/monitoring)
