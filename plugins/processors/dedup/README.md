# Dedup Processor Plugin

This plugin filters metrics whose field values are exact repetitions of the
previous values. This plugin will store its state between runs if the
`statefile` option in the agent config section is set.

⭐ Telegraf v1.14.0
🏷️ filtering
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Filter metrics with repeating field values
[[processors.dedup]]
  ## Maximum time to suppress output
  dedup_interval = "600s"
```

## Example

```diff
- cpu,cpu=cpu0 time_idle=42i,time_guest=1i
- cpu,cpu=cpu0 time_idle=42i,time_guest=2i
- cpu,cpu=cpu0 time_idle=42i,time_guest=2i
- cpu,cpu=cpu0 time_idle=44i,time_guest=2i
- cpu,cpu=cpu0 time_idle=44i,time_guest=2i
+ cpu,cpu=cpu0 time_idle=42i,time_guest=1i
+ cpu,cpu=cpu0 time_idle=42i,time_guest=2i
+ cpu,cpu=cpu0 time_idle=44i,time_guest=2i
```
