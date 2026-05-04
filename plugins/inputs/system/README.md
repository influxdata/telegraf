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
  ## Information to collect; available options are:
  ##   load             - 1, 5 and 15-minute load averages
  ##   users            - logged-in user counts
  ##   cpus             - CPU counts of the system
  ##   legacy_cpus      - legacy layout of CPU counts; see README for details
  ##   uptime           - system uptime
  ##   legacy_uptime    - legacy layout of system uptime; see README for details
  ##   os               - operating system release and uname information
  # include = ["load", "users", "legacy_cpus", "legacy_uptime"]

  ## How long to cache the result of the "os" group between gathers.
  ## Set higher to reduce the number of os-release/uname reads, lower to
  ## surface distro upgrades and kexec'd kernels faster. A value of zero
  ## ("0s") caches the values until telegraf restarts; only safe on hosts
  ## that are not re-imaged or kexec'd at runtime. To re-read on every
  ## gather, set to a very small positive value such as "1ns".
  # os_cache_ttl = "5m"
```

> [!NOTE]
> The `cpus` and `legacy_cpus` options are mutually exclusive,
> as are `uptime` and `legacy_uptime`.

<!-- markdownlint-disable-next-line MD028 -->

> [!IMPORTANT]
> Switching from `legacy_uptime` to `uptime` changes the Prometheus metric
> type of `system_uptime` from **counter** to **gauge**. If your dashboards
> or alerts use `rate()` or `increase()` on `system_uptime`, update them
> before migrating.

### Permissions

The `n_users` field requires read access to `/var/run/utmp`, and may require the
`telegraf` user to be added to the `utmp` group on some systems. If this file
does not exist `n_users` will be skipped.

The `n_unique_users` shows the count of unique usernames logged in. This way if
a user has multiple sessions open/started they would only get counted once. The
same requirements for `n_users` apply.

The `os` group reads `/etc/os-release` on Linux (typically world-readable) and
calls the `uname` syscall on POSIX systems. On platforms where gopsutil cannot
provide a particular value (e.g. parts of FreeBSD/OpenBSD/Solaris) the
corresponding field is left empty; if no field can be gathered, the
`system_os` metric is skipped entirely.

Results of the `os` group are cached for `os_cache_ttl` (default 5 minutes)
between gathers. The values rarely change at runtime, but the cache is
refreshed periodically so that distribution upgrades (which rewrite
`/etc/os-release`) and `kexec` boots (which change `uname -r`/`-m`) surface
in the metric without restarting telegraf. Set `os_cache_ttl = "0s"` to
cache the values until telegraf restarts; this is appropriate on static
hosts where the operator is sure no `kexec` boot or distribution upgrade
will occur during the agent's lifetime. To re-read the data on every
gather (effectively disabling the cache), set the TTL to a very small
positive value such as `"1ns"`.

## Metrics

The `include` option controls which measurements and fields are gathered.
The `load`, `users`, `cpus` / `legacy_cpus` and `uptime` / `legacy_uptime`
groups populate the `system` measurement, while the `os` group emits a
separate `system_os` measurement.

### `system`

| Field             | Include option             | Type    | Description                                 |
|-------------------|----------------------------|---------|---------------------------------------------|
| `load1`           | `load`                     | float   | 1-minute load average                       |
| `load5`           | `load`                     | float   | 5-minute load average                       |
| `load15`          | `load`                     | float   | 15-minute load average                      |
| `n_users`         | `users`                    | integer | Number of logged-in user sessions           |
| `n_unique_users`  | `users`                    | integer | Number of unique logged-in usernames        |
| `n_virtual_cpus`  | `cpus`                     | integer | Number of logical CPUs                      |
| `n_cpus`          | `legacy_cpus`              | integer | Number of logical CPUs (legacy name)        |
| `n_physical_cpus` | `cpus` / `legacy_cpus`     | integer | Number of physical CPUs                     |
| `uptime`          | `uptime`                   | integer | System uptime in seconds (gauge field)      |
| `uptime`          | `legacy_uptime`            | integer | System uptime in seconds (separate counter) |
| `uptime_format`   | `legacy_uptime`            | string  | Human-readable uptime (deprecated)          |

### `system_os`

Emitted only when `os` is included. The values are gathered through
[gopsutil][gopsutil] for cross-platform support and reflect operating
system release information together with `uname`-style kernel data.
Fields are reported as strings; on platforms where a particular value
cannot be determined the corresponding field is empty.

[gopsutil]: https://github.com/shirou/gopsutil

| Field              | Type   | Description                                                          |
|--------------------|--------|----------------------------------------------------------------------|
| `os`               | string | Operating system family as reported by Go's runtime (e.g. `linux`)   |
| `platform`         | string | OS distribution / platform identifier (e.g. `ubuntu`, `centos`)      |
| `platform_family`  | string | Platform family (e.g. `debian`, `rhel`)                              |
| `platform_version` | string | Platform / distribution version (e.g. `22.04`)                       |
| `kernel_version`   | string | Kernel release as returned by `uname -r` (e.g. `5.15.0-91-generic`)  |
| `kernel_arch`      | string | Kernel architecture as returned by `uname -m` (e.g. `x86_64`)        |

## Example Output

### Default configuration

With the default `include = ["load", "users", "legacy_cpus", "legacy_uptime"]`,
the output is backward-compatible with previous versions:

```text
system,host=worker-01 load1=3.72,load5=2.4,load15=2.1,n_users=3i,n_unique_users=2i,n_cpus=4i,n_physical_cpus=2i 1748000000000000000
system,host=worker-01 uptime=1249632i 1748000000000000000
system,host=worker-01 uptime_format="14 days, 11:07" 1748000000000000000
```

### Recommended configuration

With `include = ["load", "users", "cpus", "uptime"]`, all fields are emitted
in a single metric with the new field names:

```text
system,host=worker-01 load1=3.72,load5=2.4,load15=2.1,n_users=3i,n_unique_users=2i,n_virtual_cpus=4i,n_physical_cpus=2i,uptime=1249632i 1748000000000000000
```

### OS information

With `include = ["os"]`, a separate `system_os` measurement is emitted:

```text
system_os,host=worker-01 os="linux",platform="ubuntu",platform_family="debian",platform_version="22.04",kernel_version="5.15.0-91-generic",kernel_arch="x86_64" 1748000000000000000
```
