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
  ## surface distro upgrades and kexec'd kernels faster. Set to zero to
  ## re-read the data on every gather.
  # os_cache_ttl = "8h"
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
calls the `uname` syscall on POSIX systems. The `os` field is always populated
from Go's runtime, and `arch` falls back to the runtime architecture when the
kernel cannot be queried, so both are always present. On platforms where
gopsutil cannot provide platform release or kernel data (e.g. parts of
FreeBSD/OpenBSD/Solaris) the `platform`, `platform_family`, `platform_version`
and `kernel_version` fields may be empty. Results are cached between gathers,
see `os_cache_ttl` above.

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

Emitted only when `os` is included. The values reflect operating system
release information together with `uname`-style kernel data. Fields are
reported as strings. The `os` and `arch` fields are always populated; the
`platform`, `platform_family`, `platform_version` and `kernel_version` fields
may be empty on platforms where gopsutil cannot determine them.

| Field              | Type   | Description                                                          |
|--------------------|--------|----------------------------------------------------------------------|
| `os`               | string | Operating system family as reported by Go's runtime (e.g. `linux`)   |
| `arch`             | string | Architecture as returned by `uname -m` (e.g. `x86_64`)               |
| `platform`         | string | OS distribution / platform identifier (e.g. `ubuntu`, `centos`)      |
| `platform_family`  | string | Platform family (e.g. `debian`, `rhel`)                              |
| `platform_version` | string | Platform / distribution version (e.g. `26.04`)                       |
| `kernel_version`   | string | Kernel release as returned by `uname -r` (e.g. `7.0.0-7-generic`)    |

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
system_os,host=worker-01 os="linux",arch="x86_64",platform="ubuntu",platform_family="debian",platform_version="26.04",kernel_version="7.0.0-7-generic" 1748000000000000000
```
