# NATS Input Plugin

The [NATS](http://www.nats.io/about/) monitoring plugin gathers metrics from the
NATS [monitoring http server][1].

[1]: https://www.nats.io/documentation/server/gnatsd-monitoring/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Provides metrics about the state of a NATS server
# This plugin does NOT support FreeBSD
[[inputs.nats]]
  ## The address of the monitoring endpoint of the NATS server
  server = "http://localhost:8222"

  ## Maximum time to receive response
  # response_timeout = "5s"
```

## Metrics

- nats
  - tags
    - server
  - fields:
    - uptime (integer, nanoseconds)
    - mem (integer, bytes)
    - subscriptions (integer, count)
    - out_bytes (integer, bytes)
    - connections (integer, count)
    - in_msgs (integer, bytes)
    - total_connections (integer, count)
    - cores (integer, count)
    - cpu (integer, count)
    - slow_consumers (integer, count)
    - routes (integer, count)
    - remotes (integer, count)
    - out_msgs (integer, count)
    - in_bytes (integer, bytes)

## Example Output

```text
nats,server=http://localhost:8222 uptime=117158348682i,mem=6647808i,subscriptions=0i,out_bytes=0i,connections=0i,in_msgs=0i,total_connections=0i,cores=2i,cpu=0,slow_consumers=0i,routes=0i,remotes=0i,out_msgs=0i,in_bytes=0i 1517015107000000000
```
