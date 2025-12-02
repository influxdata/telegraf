# Heartbeat Output Plugin

This plugin sends a heartbeat signal via POST to a HTTP endpoint on a regular
interval. This is useful to keep track of existing Telegraf instances in a large
deployment.

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
# A plugin that can transmit heartbeats over HTTP
[[outputs.heartbeat]]
  ## URL of heartbeat endpoint
  url = "http://monitoring.example.com/heartbeat"

  ## Unique identifier to submit for the Telegraf instance (required)
  instance_id = "agent-123"

  ## Token for bearer authentication
  # token = ""

  ## Interval for sending heartbeat messages
  # interval = "1m"

  ## Information to include in the message, available options are
  ##   hostname   -- hostname of the instance running Telegraf
  ##   statistics -- number of metrics, logged errors and warnings, etc
  ##   configs    -- redacted list of configs loaded by this instance
  # include = ["hostname"]

  ## Additional HTTP headers
  # [outputs.heartbeat.headers]
  #   User-Agent = "telegraf"
```

Each heartbeat message, sent every `interval`, contains at least the specified
Telegraf `instance_id`, the Telegraf version and the version of the JSON-Schema
used for the message. The latest schema can be found in the
[plugin directory][schema].

Additional information can be included in the message via the `include` setting.

> [!NOTE]
> Some information, e.g. the number of metrics, is only updated after the first
> flush cycle, this must be considered when interpreting the messages.

Statistics included in heartbeat messages are accumulated since the last
successful heartbeat. If a heartbeat cannot be sent, accumulation of data
continues until the next successful send. Additionally, message after a failed
send the `last` field contains the Unix timestamp of the last successful
heartbeat, allowing you to identify gaps in reporting and to calculate rates.

### Configuration information

When including `configs` in the message, the heartbeat message will contain the
configuration sources used to setup the currently running Telegraf instance.

> [!WARNING]
> As the configuration sources contains the path or the URL, the resulting
> heartbeat messages may be large. Use this option with care if network
> traffic is a limiting factor!

The configuration information can potentially change when watching e.g. the
configuration directory while a new configuration is added or removed.

> [!IMPORTANT]
> Configuration URLs are redacted to remove the username and password
> information. However, sensitive information might still be contained in the
> URL or the path sent. Use with care!

[schema]: /plugins/outputs/heartbeat/schema_v1.json
