# Processes Input Plugin

This plugin gathers info about the total number of processes and groups them by
status (zombie, sleeping, running, etc.)

> [!NOTE]
> On Linux this plugin requires access to procfs (/proc), on other operating
> systems the plugin must be able to execute the `ps` command.

⭐ Telegraf v0.11.0
🏷️ system
💻 freebsd, linux, macos

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Get the number of processes and group them by status
# This plugin ONLY supports non-Windows
[[inputs.processes]]
  ## Use sudo to run ps command on *BSD systems. Linux systems will read
  ## /proc, so this does not apply there.
  # use_sudo = false
```

Another possible configuration is to define an alternative path for resolving
the /proc location.  Using the environment variable `HOST_PROC` the plugin will
retrieve process information from the specified location.

`docker run -v /proc:/rootfs/proc:ro -e HOST_PROC=/rootfs/proc`

### Using sudo

Linux systems will read from `/proc`, while BSD systems will use the `ps`
command. The `ps` command generally does not require elevated permissions.
However, if a user wants to collect system-wide stats, elevated permissions are
required. If the user has configured sudo with the ability to run this
command, then set the `use_sudo` to true.

If your account does not already have the ability to run commands with
passwordless sudo then updates to the sudoers file are required. Below is an
example to allow the requires ps commands:

First, use the `visudo` command to start editing the sudoers file. Then add
the following content, where `<username>` is the username of the user that
needs this access:

```text
Cmnd_Alias PS = /bin/ps
<username> ALL=(root) NOPASSWD: PS
Defaults!PS !logfile, !syslog, !pam_session
```

## Metrics

- processes
  - fields:
    - blocked (aka disk sleep or uninterruptible sleep)
    - running
    - sleeping
    - stopped
    - total
    - zombie
    - dead
    - wait (freebsd only)
    - idle (bsd and Linux 4+ only)
    - paging (linux only)
    - parked (linux only)
    - total_threads (linux only)

Different OSes use slightly different State codes for their processes, these
state codes are documented in `man ps`, and I will give a mapping of what major
OS state codes correspond to in telegraf metrics:

```text
Linux  FreeBSD  Darwin  meaning
  R       R       R     running
  S       S       S     sleeping
  Z       Z       Z     zombie
  X      none    none   dead
  T       T       T     stopped
  I       I       I     idle (sleeping for longer than about 20 seconds)
  D      D,L      U     blocked (waiting in uninterruptible sleep, or locked)
  W       W      none   paging (linux kernel < 2.6 only), wait (freebsd)
```

## Example Output

```text
processes blocked=8i,running=1i,sleeping=265i,stopped=0i,total=274i,zombie=0i,dead=0i,paging=0i,total_threads=687i 1457478636980905042
```
