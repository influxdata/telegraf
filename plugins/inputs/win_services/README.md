# Windows Services Input Plugin

This plugin collects information about the status of Windows services.

> [!NOTE]
> Monitoring some services may require running Telegraf with administrator
> privileges.

⭐ Telegraf v1.4.0
🏷️ system
💻 windows

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Input plugin to report Windows services info.
# This plugin ONLY supports Windows
[[inputs.win_services]]
  ## Names of the services to monitor. Leave empty to monitor all the available
  ## services on the host. Globs accepted. Case insensitive.
  service_names = [
    "LanmanServer",
    "TermService",
    "Win*",
  ]

  # optional, list of service names to exclude
  excluded_service_names = ['WinRM']
```

## Metrics

- win_services
  - tags
    - service_name
    - display_name
  - fields
    - state (integer)
    - startup_mode (integer)

The `state` field can have the following values:

- `1` - stopped
- `2` - start pending
- `3` - stop pending
- `4` - running
- `5` - continue pending
- `6` - pause pending
- `7` - paused

The `startup_mode` field can have the following values:

- `0` - boot start
- `1` - system start
- `2` - auto start
- `3` - demand start
- `4` - disabled

## Example Output

```text
win_services,host=WIN2008R2H401,display_name=Server,service_name=LanmanServer state=4i,startup_mode=2i 1500040669000000000
win_services,display_name=Remote\ Desktop\ Services,service_name=TermService,host=WIN2008R2H401 state=1i,startup_mode=3i 1500040669000000000
```
