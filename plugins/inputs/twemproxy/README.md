# Twemproxy Input Plugin

This plugin gathers statistics from [Twemproxy][twemproxy] servers.

⭐ Telegraf v0.3.0
🏷️ server
💻 all

[twemproxy]: https://github.com/twitter/twemproxy

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read Twemproxy stats data
[[inputs.twemproxy]]
  ## Twemproxy stats address and port (no scheme)
  addr = "localhost:22222"
  ## Monitor pool name
  pools = ["redis_pool", "mc_pool"]
```

## Metrics

- twemproxy
  - tags
    - source
    - twemproxy
  - fields
    - curr_connections (float)
    - timestamp (float)
    - total_connections (float)

- twemproxy_pool
  - tags
    - pool
    - source
    - twemproxy
  - fields
    - client_connections (float)
    - client_eof (float)
    - client_err (float)
    - forward_error (float)
    - fragments (float)
    - server_ejects (float)

- twemproxy_pool_server
  - tags
    - pool
    - server
    - source
    - twemproxy
  - fields
    - in_queue (float)
    - in_queue_bytes (float)
    - out_queue (float)
    - out_queue_bytes (float)
    - requests (float)
    - request_bytes (float)
    - responses (float)
    - response_bytes (float)
    - server_connections (float)
    - server_ejected_at (float)
    - server_eof (float)
    - server_err (float)
    - server_timedout (float)

## Example Output

```text
twemproxy,source=server1.website.com,twemproxy=127.0.0.1:22222 curr_connections=1322,timestamp=1447312436,total_connections=276448 1748893350082522719
twemproxy_pool_server,pool=demo,server=10.16.29.1:6379,source=server1.website.com,twemproxy=127.0.0.1:22222 in_queue=0,in_queue_bytes=0,out_queue=0,out_queue_bytes=0,request_bytes=2775840400,requests=43604566,response_bytes=7663182096,responses=43603900,server_connections=1,server_ejected_at=0,server_eof=0,server_err=0,server_timedout=24 1748893350082546069
twemproxy_pool_server,pool=demo,server=10.16.29.2:6379,source=server1.website.com,twemproxy=127.0.0.1:22222 in_queue=0,in_queue_bytes=0,out_queue=0,out_queue_bytes=0,request_bytes=2412114759,requests=37870211,response_bytes=5228980582,responses=37869551,server_connections=1,server_ejected_at=0,server_eof=0,server_err=0,server_timedout=25 1748893350082560329
twemproxy_pool,pool=demo,source=server1.website.com,twemproxy=127.0.0.1:22222 client_connections=1305,client_eof=126813,client_err=147942,forward_error=11684,fragments=0,server_ejects=0 1748893350082572369
```
