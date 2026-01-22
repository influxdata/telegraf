# Raindrops Middleware Input Plugin

This plugin collects statistics for [Raindrops middleware][raindrops] instances.

⭐ Telegraf v0.10.3
🏷️ server
💻 all

[raindrops]: http://raindrops.bogomips.org/Raindrops/Middleware.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

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
  - tags:
    - server
    - port
  - fields:
    - calling (integer, count)
    - writing (integer, count)
- raindrops_listen
  - tags:
    - ip   (IP only)
    - port (IP only)
    - socket (unix socket only)
  - fields:
    - active (integer, bytes)
    - queued (integer, bytes)

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
