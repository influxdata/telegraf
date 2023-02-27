# Radius Input Plugin

The Radius plugin collects radius authentication response times.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
[[inputs.radius]]
  ## An array of Server IPs and ports to gather from. If none specified, defaults to localhost.
  servers = ["127.0.0.1:1812","hostname.domain.com:1812"]

  ## Credentials for radius authentication.
  username = "myuser"
  password = "mypassword"
  secret = "mysecret"

  ## Maximum time to receive response.
  # response_timeout = "5s"
```

## Metrics

- radius
  - tags:
    - port
    - source
  - fields:
    - responsetime (float)

## Example Output

```shell
radius,port=1812,source=debian-stretch-radius responsetime=0.011 1502489900000000000
```
