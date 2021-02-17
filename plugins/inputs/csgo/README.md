# CSGO Input Plugin

The `csgo` plugin gather metrics from CSGO servers.

#### Configuration
```toml
[[inputs.csgo]]
  servers = [
    ["ip1:port1", "rcon_password1"],
    ["ip2:port2", "rcon_password2"],
  ]
```

### Metrics

The plugin retrieves the output of the `stats` command that is executed via rcon.

If no servers are specified, no data will be collected

- csgo
    - tags:
        - host
    - fields:
        - csgo_cpu (float)
        - csgo_net_in (float)
        - csgo_net_out (float)
        - csgo_uptime_minutes (float)
        - csgo_maps (float)
        - csgo_fps (float)
        - csgo_players (float)
        - csgo_svms (float)
        - csgo_ms_var (float)
        - csgo_tick (float)
