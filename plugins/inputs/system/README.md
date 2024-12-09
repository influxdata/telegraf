# System Input Plugin

The system plugin gathers general stats on system load, uptime,
and number of users logged in. It is similar to the unix `uptime` command.

Number of CPUs is obtained from the /proc/cpuinfo file.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about system load & uptime
[[inputs.system]]
  # no configuration
```

### Permissions

The `n_users` field requires read access to `/var/run/utmp`, and may require the
`telegraf` user to be added to the `utmp` group on some systems. If this file
does not exist `n_users` will be skipped.

The `n_unique_users` shows the count of unique usernames logged in. This way if
a user has multiple sessions open/started they would only get counted once. The
same requirements for `n_users` apply.

## Metrics

- system
  - fields:
    - load1 (float)
    - load15 (float)
    - load5 (float)
    - n_users (integer)
    - n_unique_users (integer)
    - n_cpus (integer)
    - uptime (integer, seconds)
    - uptime_format (string, deprecated in 1.10, use `uptime` field)

## Example Output

```text
system,host=tyrion load1=3.72,load5=2.4,load15=2.1,n_users=3i,n_cpus=4i 1483964144000000000
system,host=tyrion uptime=1249632i 1483964144000000000
system,host=tyrion uptime_format="14 days, 11:07" 1483964144000000000
```
