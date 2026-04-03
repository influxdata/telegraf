# System Input Plugin

This plugin gathers general system statistics like system load, uptime or the
number of users logged in. It is similar to the unix `uptime` command.

⭐ Telegraf v0.1.6
🏷️ system
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about system load & uptime
[[inputs.system]]
  ## Metric groups to collect.
  ## Available options:
  ##   load   - system gauge metrics (load averages, cpu counts, user counts)
  ##   uptime - system uptime
  ## By default all groups are collected.
  # collect = ["load", "uptime"]
```

### Permissions

The `n_users` field requires read access to `/var/run/utmp`, and may require the
`telegraf` user to be added to the `utmp` group on some systems. If this file
does not exist `n_users` will be skipped.

The `n_unique_users` shows the count of unique usernames logged in. This way if
a user has multiple sessions open/started they would only get counted once. The
same requirements for `n_users` apply.

## Metrics

### `system`

All fields below belong to the `system` measurement. The `collect` option
controls which groups are gathered.

| Field             | Group    | Type    | Description                                    |
|-------------------|----------|---------|------------------------------------------------|
| `load1`           | `load`   | float   | 1-minute load average                          |
| `load5`           | `load`   | float   | 5-minute load average                          |
| `load15`          | `load`   | float   | 15-minute load average                         |
| `n_users`         | `load`   | integer | Number of logged-in user sessions              |
| `n_unique_users`  | `load`   | integer | Number of unique logged-in usernames           |
| `n_cpus`          | `load`   | integer | Number of logical CPUs                         |
| `n_physical_cpus` | `load`   | integer | Number of physical CPUs                        |
| `uptime`          | `uptime` | integer | System uptime in seconds                       |
| `uptime_format`   | `uptime` | string  | Human-readable uptime (deprecated, use uptime) |

## Example Output

```text
system,host=worker-01 load1=3.72,load5=2.4,load15=2.1,n_users=3i,n_unique_users=2i,n_cpus=4i,n_physical_cpus=2i 1748000000000000000
system,host=worker-01 uptime=1249632i 1748000000000000000
system,host=worker-01 uptime_format="14 days, 11:07" 1748000000000000000
```

## Example Output (Prometheus)

When using the [Prometheus output plugin][prom-output] or
[Prometheus client plugin][prom-client], Telegraf converts each field into
its own Prometheus metric by appending the field name to the measurement name.

[prom-output]: ../../../plugins/outputs/prometheus_client/README.md
[prom-client]: ../../../plugins/outputs/prometheus_client/README.md

```text
# HELP system_load1 Telegraf collected metric
# TYPE system_load1 gauge
system_load1{host="worker-01"} 3.72

# HELP system_load15 Telegraf collected metric
# TYPE system_load15 gauge
system_load15{host="worker-01"} 2.1

# HELP system_load5 Telegraf collected metric
# TYPE system_load5 gauge
system_load5{host="worker-01"} 2.4

# HELP system_n_cpus Telegraf collected metric
# TYPE system_n_cpus gauge
system_n_cpus{host="worker-01"} 4

# HELP system_n_physical_cpus Telegraf collected metric
# TYPE system_n_physical_cpus gauge
system_n_physical_cpus{host="worker-01"} 2

# HELP system_n_unique_users Telegraf collected metric
# TYPE system_n_unique_users gauge
system_n_unique_users{host="worker-01"} 2

# HELP system_n_users Telegraf collected metric
# TYPE system_n_users gauge
system_n_users{host="worker-01"} 3

# HELP system_uptime Telegraf collected metric
# TYPE system_uptime counter
system_uptime{host="worker-01"} 1249632
```
