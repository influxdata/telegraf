# Heartbeat Output Plugin

This plugin sends a heartbeat signal via POST to a HTTP endpoint on a regular
interval. This is useful to keep track of existing Telegraf instances in a large
deploayment.

⭐ Telegraf v1.37.0
🏷️ applications
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `url`, `token` and
`headers` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# A plugin that can transmit metrics over HTTP
[[outputs.heartbeat]]
  ## URL of heartbeat endpoint
  url = "http://monitoring.example.com/heartbeat"

  ## Unique identifier to submit for the Telegraf instance (required)
  instance_id = "agent-123"

  ## Token for bearer authentication
  # token = ""

  ## Interval for sending heartbeart messages
  # interval = "1m"

  ## Information to include in the message, available options are
  ##   hostname -- hostname of the instance running Telegraf
  ##   metrics  -- number of metrics added during the last interval
  ##   configs  -- redacted list of configs loaded by this instance
  ##   logs     -- log-entries for this instance
  ##   status   -- result of the status condition evaluation
  # include = ["hostname", "status"]

  ## Logical conditions to determine the agent status, only applies if "status"
  ## is included in the message
  # status_ok = "true"

  ## Logging information filtering, only applies if "logs" is added to "include"
  # [outputs.heartbeat.logs]
  #   ## Number of log entries to send (unlimited by default)
  #   ## In case more log-entries are available entires with higher log levels
  #   ## and more recent entires are preferred.
  #   # limit = 0
  #
  #   ## Minimum log-level for sending the entry
  #   # level = "error"

  ## Additional HTTP headers
  # [outputs.heartbeat.headers]
  #   User-Agent = "telegraf"
```
