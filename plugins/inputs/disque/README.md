# Disque Input Plugin

This plugin gathers data from a [Disque][disque] instance, an experimental
distributed, in-memory, message broker.

⭐ Telegraf v0.10.0
🏷️ messaging
💻 all

[disque]: https://github.com/antirez/disque

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from one or many disque servers
[[inputs.disque]]
  ## An array of URI to gather stats about. Specify an ip or hostname
  ## with optional port and password.
  ## ie disque://localhost, disque://10.10.3.33:18832, 10.0.0.1:10000, etc.
  ## If no servers are specified, then localhost is used as the host.
  servers = ["localhost"]
```

## Metrics

- disque
  - disque_host
    - uptime_in_seconds
    - connected_clients
    - blocked_clients
    - used_memory
    - used_memory_rss
    - used_memory_peak
    - total_connections_received
    - total_commands_processed
    - instantaneous_ops_per_sec
    - latest_fork_usec
    - mem_fragmentation_ratio
    - used_cpu_sys
    - used_cpu_user
    - used_cpu_sys_children
    - used_cpu_user_children
    - registered_jobs
    - registered_queues

## Example Output
