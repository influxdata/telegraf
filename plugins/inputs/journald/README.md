# Systemd Journal Input Plugin

This service plugin continuously reads entries from the [systemd journal][journal]
by following the JSON output of [journalctl][journalctl] and emits one metric
per entry. Entries can be filtered by unit, priority and arbitrary journal
matches. The position in the journal is persisted, so no entries are lost or
duplicated across restarts if a `statefile` is configured for the agent.

⭐ Telegraf v1.41.0
🏷️ logging, system
💻 linux

[journal]: https://www.freedesktop.org/software/systemd/man/latest/systemd-journald.service.html
[journalctl]: https://www.freedesktop.org/software/systemd/man/latest/journalctl.html

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listen and wait for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Tracking metric support <!-- @/docs/includes/plugin_tracking_metrics.md -->

This plugin supports [tracking metrics][METRICS.md], which allows the plugin
to be notified when metrics have been delivered to all outputs, enabling proper
acknowledgment back to the source.

[METRICS.md]: ../../../docs/METRICS.md#tracking-metrics

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read log entries from the systemd journal
# This plugin ONLY supports Linux
[[inputs.journald]]
  ## Use the specified binary which will be looked-up in PATH
  # binary = "journalctl"

  ## Use sudo for command execution, the telegraf user can alternatively be
  ## added to the "systemd-journal" group to read the system journal
  # use_sudo = false

  ## Only read entries of the given systemd system or user units, glob
  ## patterns are supported (e.g. "nginx.service" or "docker-*")
  # units = []
  # user_units = []

  ## Only read entries with the given priority or a more important one,
  ## either a single value (e.g. "warning" or "4") or a range ("err..notice")
  # priority = ""

  ## Additional journal matches passed to journalctl, see journalctl(1),
  ## e.g. ["_TRANSPORT=kernel"]
  # matches = []

  ## Read the journal files in the given directory or the given journal
  ## namespace instead of the default system journal
  # directory = ""
  # namespace = ""

  ## Position to start reading at
  ## The following methods are available:
  ##   beginning          -- start at the oldest entry ignoring any persisted cursor
  ##   end                -- only read new entries ignoring any persisted cursor
  ##   saved-or-beginning -- continue after the persisted cursor or, if no
  ##                         cursor is persisted, start at the oldest entry
  ##   saved-or-end       -- continue after the persisted cursor or, if no
  ##                         cursor is persisted, only read new entries
  ## The cursor is only persisted if a "statefile" is configured in the agent
  ## section and it only advances once entries were written by the outputs.
  # initial_read_offset = "saved-or-end"

  ## Journal fields to add as tags, the field names are lowercased and leading
  ## underscores are removed (e.g. "_BOOT_ID" becomes the "boot_id" tag)
  # tag_fields = []

  ## Journal fields to add as string fields, glob patterns are supported
  ## (e.g. ["CODE_*", "_EXE"]), the names are converted like for tags
  # journal_fields = []

  ## Maximum number of entries that have not yet been written by the outputs.
  ## For best throughput set it to at least the metric_batch_size of the
  ## outputs.
  # max_undelivered_entries = 1000

  ## Time to wait before restarting journalctl if it exits unexpectedly
  # restart_delay = "5s"
```

### Permissions

The user running Telegraf needs read access to the journal. Either add the
`telegraf` user to the `systemd-journal` group (recommended) or enable
`use_sudo` and allow the command in your sudoers configuration, e.g.

```text
telegraf ALL=(root) NOPASSWD: /usr/bin/journalctl
```

Without access, journalctl only shows the entries of the `telegraf` user and
prints a warning that is forwarded to the Telegraf log.

### Persisting the position

The plugin implements the Telegraf state interface and saves the [cursor][]
of the last entry that was written by all outputs. With a `statefile`
configured in the `[agent]` section and `initial_read_offset` set to
`saved-or-beginning` or `saved-or-end`, the plugin continues right after that
entry when Telegraf restarts, so entries are neither lost nor read twice. If
journalctl rejects the persisted cursor (e.g. because the journal was rotated
or vacuumed), the plugin logs a warning and starts according to
`initial_read_offset`.

If journalctl exits while Telegraf is running, it is restarted after
`restart_delay` and continues after the last entry read.

[cursor]: https://www.freedesktop.org/software/systemd/man/latest/sd_journal_get_cursor.html

## Metrics

- journald
  - tags:
    - hostname (string) - Host that logged the entry (`_HOSTNAME`)
    - unit (string, optional) - Systemd unit of the process (`_SYSTEMD_UNIT`)
    - user_unit (string, optional) - Systemd user unit of the process
      (`_SYSTEMD_USER_UNIT`)
    - identifier (string, optional) - Syslog identifier of the process
      (`SYSLOG_IDENTIFIER`)
    - severity (string, optional) - Syslog severity name, e.g. `err` or `info`
      (`PRIORITY`)
    - facility (string, optional) - Syslog facility name, e.g. `daemon`
      (`SYSLOG_FACILITY`)
    - additional tags configured in `tag_fields`
  - fields:
    - message (string) - Message of the entry (`MESSAGE`)
    - severity_code (int, optional) - Syslog severity code (`PRIORITY`)
    - facility_code (int, optional) - Syslog facility code (`SYSLOG_FACILITY`)
    - pid (int, optional) - Process ID (`_PID`)
    - additional string fields configured in `journal_fields`

The metric timestamp is the time the entry was received by the journal
(`__REALTIME_TIMESTAMP`). Binary field values are converted to strings and
for fields occurring multiple times in an entry, the first value is used.

## Example Output

```text
journald,facility=auth,hostname=web-01,identifier=sshd,severity=info,unit=ssh.service facility_code=4i,message="Accepted publickey for deploy from 10.0.0.12 port 51234 ssh2",pid=1734i,severity_code=6i 1762265101000000000
journald,facility=kern,hostname=web-01,identifier=kernel,severity=warning facility_code=0i,message="EXT4-fs (sda1): warning: mounting fs with errors, running e2fsck is recommended",severity_code=4i 1762265101250001000
journald,facility=daemon,hostname=web-01,identifier=backup,severity=err,unit=user@1000.service,user_unit=backup.service facility_code=3i,message="upload of \"db.tar\" failed: connection reset",pid=2211i,severity_code=3i 1762265101500002000
```
