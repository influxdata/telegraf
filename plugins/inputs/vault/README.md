# Hashicorp Vault Input Plugin

This plugin collects metrics from every [Vault][vault] agent of a cluster.

> [!NOTE]
> This plugin requires Vault v1.8.5+

⭐ Telegraf v1.22.0
🏷️ server
💻 all

[vault]: https://www.hashicorp.com/de/products/vault

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
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

For a more deep understanding of Vault monitoring, please have a look at the
following Vault [telemetry][telemetry] and [monitoring][monitoring]
documentation.

[telemetry]: https://www.vaultproject.io/docs/internals/telemetry
[monitoring]: https://learn.hashicorp.com/tutorials/vault/monitor-telemetry-audit-splunk?in=vault/monitoring

## Example Output

```text
vault.raft.replication.appendEntries.logs,peer_id=clustnode-02 count=130i,max=1i,mean=0.015384615384615385,min=0i,rate=0.2,stddev=0.12355304447984486,sum=2i 1638287340000000000
vault.core.unsealed,cluster=vault-cluster-23b671c7 value=1i 1638287340000000000
vault.token.lookup count=5135i,max=16.22449493408203,mean=0.1698389152269865,min=0.06690400093793869,rate=87.21228296905755,stddev=0.24637634000854705,sum=872.1228296905756 1638287340000000000
```
