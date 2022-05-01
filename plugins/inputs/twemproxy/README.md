# Twemproxy Input Plugin

The `twemproxy` plugin gathers statistics from [Twemproxy](https://github.com/twitter/twemproxy) servers.

## Configuration

```toml
# Read Twemproxy stats data
[[inputs.twemproxy]]
  ## Twemproxy stats address and port (no scheme)
  addr = "localhost:22222"
  ## Monitor pool name
  pools = ["redis_pool", "mc_pool"]
```
