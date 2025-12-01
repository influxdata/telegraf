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
  ##   logs       -- detailed log-entries for this instance
  ##   status     -- result of the status condition evaluation
  # include = ["hostname", "status"]

  ## Logging information filtering, only applies if "logs" is added to "include"
  # [outputs.heartbeat.logs]
  #   ## Number of log entries to send (unlimited by default)
  #   ## In case more log-entries are available entires with higher log levels
  #   ## and more recent entires are preferred.
  #   # limit = 0
  #
  #   ## Minimum log-level for sending the entry
  #   # level = "error"

  ## Logical conditions to determine the agent status, only applies if "status"
  ## is included in the message
  # [outputs.heartbeat.status]
  #   ## Conditions to signal the given status as CEL programs returning a
  #   ## boolean. Conditions are evaluated in the order below until a program
  #   ## evaluates to "true".
  #   # ok = "false"
  #   # warn = "false"
  #   # fail = "false"
  #
  #   ## Evaluation order of the conditions above; available: "ok", "warn", "fail"
  #   # order = ["ok", "warn", "fail"]
  #
  #   ## Default status used if none of the conditions above matches
  #   ## available: "ok", "warn", "fail", "undefined"
  #   # default = "ok"
  #
  #   ## If set, send this initial status before the first write, otherwise
  #   ## compute the status from the conditions and default above.
  #   ## available: "ok", "warn", "fail", "undefined", ""
  #   # initial = ""

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

When including `configs` in the message the heartbeat message will contain the
configuration sources used to setup the currently running Telegraf instance.

> [!WARNING]
> As the configuration sources contains path or the URL the resulting heartbeat
> messages might be large. Use this option with care if network traffic is a
> limiting factor!

The configuration information can potentially change when watching e.g. the
configuration directory while a new configuration is added or removed.

> [!IMPORTANT]
> Configuration URLs are redacted to remove the username and password
> information. However, sensitive information might still be contained in the
> URL or the path sent. Use with care!

### Logging information

When including `logs` in the message the number of errors and warnings logged
in this Telegraf instance are included in the heartbeat message. This comprises
_all_ log messages of all plugins and the agent itself logged _after_ the
`Connect` function of this plugin was called, i.e. you will not see any
initialization or configuration errors in the heartbeat messages!

For getting the actual log _messages_ you can include `log-details`. Via the
optional `outputs.heartbeat.status` you can limit the messages by log-`level`
or limit the number included using the `limit` setting.

> [!WARNING]
> As the amount of log messages can be high, especially when configuring low
> level such as `info` the resulting heartbeat messages might be large. Use the
> `log-details` option with care if network traffic is a limiting factor and
> restrict the included messages to high levels and use a limit!

When setting the `level` option only messages with this or more severe levels
are included.

The `limit` setting allows to specify the maximum number of log-messages
included in the heartbeat message. If the number of log-messages exceeds the
given limit they are selected by most severe and most recent messages first.

### Status information

By including `status` the message will contain the status of the Telegraf
instance as configured via the `outputs.heartbeat.status` section.

This section allows to set an `initial` state used as long as no flush was
performed by Telegraf. If `initial` is not configured or empty, the status
expressions are evaluated also before the first flush.

The `ok`, `warn` and `fail` settings allow to specify [CEL expressions][cel]
evaluating to a boolean value. Available information for the expressions are
listed below. The first expression evaluating to `true` defines the status.
The `order` parameter allows to customize the evaluation order.

> [!NOTE]
> If an expression is omitted in the `order` setting it will __not__ be
> evaluated!

The status defined via `default` is used in case none of the status expressions
evaluate to true.

For defining expressions you can use the following variables

- `metrics` (int)      -- number of metrics arriving at this plugin
- `log_errors` (int)   -- number of errors logged
- `log_warnings` (int) -- number of warnings logged
- `last_update` (int)  -- time of last successful heartbeat message, can be used
                          to e.g. calculate rates
- `agent` (map)        -- agent statistics, see below
- `inputs` (map)       -- input plugin statistics, see below
- `outputs` (map)      -- output plugin statistics, see below

The `agent` statistics variable is a `map` with information matching the
`internal_agent` metric of the [internal input plugin][internal_plugin]:

- `metrics_written` (int)  -- number of metrics written by the agent
- `metrics_rejected` (int) -- number of metrics rejected by the agent
- `metrics_dropped` (int)  -- number of metrics dropped by the agent

The `inputs` statistics variable is a `map` with the key denoting the plugin
type (e.g. `cpu` for `inputs.cpu`) and the value being list of plugin
statistics. Each entry in the list corresponds to an input plugin instance with
information matching the `internal_gather` metric of the
[internal input plugin][internal_plugin]:

- `id` (string)            -- unique plugin identifier
- `alias` (string)         -- alias set for the plugin; only exists if alias
                              is defined
- `errors` (int)           -- collection errors for this plugin instance
- `metrics_gathered` (int) -- number of metrics collected
- `gather_time_ns` (int)   -- time used to gather the metrics in nanoseconds
- `gather_timeouts` (int)  -- number of timeouts during metric collection
- `startup_errors` (int)   -- number of times the plugin failed to start

The `outputs` statistics variable is a `map` with the key denoting the plugin
type (e.g. `influxdb` for `outputs.influxdb`) and the value being list of plugin
statistics. Each entry in the list corresponds to an output plugin instance with
information matching the `internal_write` metric of the
[internal input plugin][internal_plugin]:

- `id` (string)             -- unique plugin identifier
- `alias` (string)          -- alias set for the plugin; only exists if alias
                               is defined
- `errors` (int)            -- write errors for this plugin instance
- `metrics_filtered` (int)  -- number of metrics filtered by the output
- `write_time_ns` (int)     -- time used to write the metrics in nanoseconds
- `startup_errors` (int)    -- number of times the plugin failed to start
- `metrics_added` (int)     -- number of metrics added to the output buffer
- `metrics_written` (int)   -- number of metrics written to the output
- `metrics_rejected` (int)  -- number of metrics rejected by the service or
                               serialization
- `metrics_dropped` (int)   -- number of metrics dropped e.g. due to buffer
                               fullness
- `buffer_size` (int)       -- current number of metrics currently in the output
                               buffer for the plugin instance
- `buffer_limit` (int)      -- capacity of the output buffer; irrelevant for
                               disk-based buffers
- `buffer_fullness` (float) -- current ratio of metrics in the buffer to
                               capacity; can be greater than one (i.e. `> 100%`)
                               for disk-based buffers

If not stated otherwise, all variables are accumulated since the last successful
heartbeat message.

[schema]: /plugins/outputs/heartbeat/schema_v1.json
[cel]: https://cel.dev
[internal_plugin]: /plugins/inputs/internal/README.md
