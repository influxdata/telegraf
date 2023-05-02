# Tacacs Input Plugin

The Tacacs plugin collects tacacs authentication response times.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
[[inputs.tacacs]]
  ## An array of Server IPs and ports to gather from. If none specified, defaults to localhost.
  # servers = ["127.0.0.1:49"]

  ## Request source server IP, normally the server running telegraf.
  # request_ip = "127.0.0.1"

  ## Credentials for tacacs authentication.
  username = "myuser"
  password = "mypassword"
  secret = "mysecret"

  ## Maximum time to receive response.
  # response_timeout = "5s"
```

## Metrics

- tacacs
  - tags:
    - source
  - fields:
    - responsetime_ms (int64)

## Example Output

```text
tacacs,source=127.0.0.1:49 responsetime_ms=311i 1677526200000000000
```
