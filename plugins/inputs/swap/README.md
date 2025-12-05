# Swap Input Plugin

This plugin collects metrics on the operating-system's swap memory.

⭐ Telegraf v1.7.0
🏷️ system
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about swap memory usage
[[inputs.swap]]
  # no configuration
```

## Metrics

- swap
  - fields:
    - free (int, bytes): free swap memory
    - total (int, bytes): total swap memory
    - used (int, bytes): used swap memory
    - used_percent (float, percent): percentage of swap memory used
    - in (int, bytes): data swapped in since last boot calculated from page number
    - out (int, bytes): data swapped out since last boot calculated from page number

## Example Output

```text
swap total=20855394304i,used_percent=45.43883523785713,used=9476448256i,free=1715331072i 1511894782000000000
```
