# Counter-Strike: Global Offensive (CSGO) Input Plugin

This plugin gather metrics from [Counter-Strike: Global Offensive][csgo]
servers.

⭐ Telegraf v1.18.0
🏷️ server
💻 all

[csgo]: https://www.counter-strike.net/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Fetch metrics from a CSGO SRCDS
[[inputs.csgo]]
  ## Specify servers using the following format:
  ##    servers = [
  ##      ["ip1:port1", "rcon_password1"],
  ##      ["ip2:port2", "rcon_password2"],
  ##    ]
  #
  ## If no servers are specified, no data will be collected
  servers = []
```

## Metrics

The plugin retrieves the output of the `stats` command that is executed via
rcon.

If no servers are specified, no data will be collected

- csgo
  - tags:
    - host
  - fields:
    - cpu (float)
    - net_in (float)
    - net_out (float)
    - uptime_minutes (float)
    - maps (float)
    - fps (float)
    - players (float)
    - sv_ms (float)
    - variance_ms (float)
    - tick_ms (float)

## Example Output
