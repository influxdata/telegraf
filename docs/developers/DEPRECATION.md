# Deprecation

Deprecation is the primary tool for making changes in Telegraf.  A deprecation
indicates that the community should move away from using a feature, and
documents that the feature will be removed in the next major update (2.0).

Key to deprecation is that the feature remains in Telegraf and the behavior is
not changed.

We do not have a strict definition of a breaking change.  All code changes
change behavior, the decision to deprecate or make the change immediately is
decided based on the impact.

## Deprecate plugins

Add an entry to the plugins deprecation list (e.g. in `plugins/inputs/deprecations.go`). Include the deprecation version
and any replacement, e.g.

```golang
  "logparser": {
    Since:  "1.15.0",
    Notice: "use 'inputs.tail' with 'grok' data format instead",
  },
```

The entry can contain an optional `RemovalIn` field specifying the planned version for removal of the plugin.

Also add the deprecation warning to the plugin's README:

```markdown
# Logparser Input Plugin

### **Deprecated in 1.10**: Please use the [tail][] plugin along with the
`grok` [data format][].

[tail]: /plugins/inputs/tail/README.md
[data formats]: /docs/DATA_FORMATS_INPUT.md
```

Telegraf will automatically check if a deprecated plugin is configured and print a warning

```text
2022-01-26T20:08:15Z W! DeprecationWarning: Plugin "inputs.logparser" deprecated since version 1.15.0 and will be removed in 2.0.0: use 'inputs.tail' with 'grok' data format instead
```

## Deprecate options

Mark the option as deprecated in the sample config, include the deprecation
version and any replacement.

```toml
  ## Broker URL
  ##   deprecated in 1.7; use the brokers option
  # url = "amqp://localhost:5672/influxdb"
```

In the plugins configuration struct, add a `deprecated` tag to the option:

```go
type AMQPConsumer struct {
    URL string `toml:"url" deprecated:"1.7.0;use brokers"`
}
```

The `deprecated` tag has the format `<since version>[;removal version];<notice>` where the `removal version` is optional. The specified deprecation info will automatically displayed by Telegraf if the option is used in the config

```text
2022-01-26T20:08:15Z W! DeprecationWarning: Option "url" of plugin "inputs.amqp_consumer" deprecated since version 1.7.0 and will be removed in 2.0.0: use brokers
```

## Deprecate metrics

In the README document the metric as deprecated.  If there is a replacement field,
tag, or measurement then mention it.

```markdown
- system
  - fields:
    - uptime_format (string, deprecated in 1.10: use `uptime` field)
```

Add filtering to the sample config, leave it commented out.

```toml
[[inputs.system]]
  ## Uncomment to remove deprecated metrics.
  # fielddrop = ["uptime_format"]
```
