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
  servers = ["127.0.0.1:49","hostname.domain.com:49"]

  ## Request source server IP, normally the server running telegraf.
  remaddr = "127.0.0.1"

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
    - responsetime (float)

## Example Output

```shell
tacacs,source=debian-stretch-tacacs responsetime=0.011 1502489900000000000
```
