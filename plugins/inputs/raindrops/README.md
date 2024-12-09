# Raindrops Input Plugin

The [raindrops](http://raindrops.bogomips.org/) plugin reads from specified
raindops [middleware](http://raindrops.bogomips.org/Raindrops/Middleware.html)
URI and adds stats to InfluxDB.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read raindrops stats (raindrops - real-time stats for preforking Rack servers)
[[inputs.raindrops]]
  ## An array of raindrops middleware URI to gather stats.
  urls = ["http://localhost:8080/_raindrops"]
```

## Metrics

- raindrops
  - calling (integer, count)
  - writing (integer, count)
- raindrops_listen
  - active (integer, bytes)
  - queued (integer, bytes)

### Tags

- Raindops calling/writing of all the workers:
  - server
  - port

- raindrops_listen (ip:port):
  - ip
  - port

- raindrops_listen (Unix Socket):
  - socket

## Example Output

```text
raindrops,port=8080,server=localhost calling=0i,writing=0i 1455479896806238204
raindrops_listen,ip=0.0.0.0,port=8080 active=0i,queued=0i 1455479896806561938
raindrops_listen,ip=0.0.0.0,port=8081 active=1i,queued=0i 1455479896806605749
raindrops_listen,ip=127.0.0.1,port=8082 active=0i,queued=0i 1455479896806646315
raindrops_listen,ip=0.0.0.0,port=8083 active=0i,queued=0i 1455479896806683252
raindrops_listen,ip=0.0.0.0,port=8084 active=0i,queued=0i 1455479896806712025
raindrops_listen,ip=0.0.0.0,port=3000 active=0i,queued=0i 1455479896806779197
raindrops_listen,socket=/tmp/listen.me active=0i,queued=0i 1455479896806813907
```
