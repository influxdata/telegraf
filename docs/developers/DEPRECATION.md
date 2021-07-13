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

Add a comment to the plugin's sample config, include the deprecation version
and any replacement.

```toml
[[inputs.logparser]]
  ## DEPRECATED: The 'logparser' plugin is deprecated in 1.10.  Please use the
  ## 'tail' plugin with the grok data_format as a replacement.
```

Add the deprecation warning to the plugin's README:

```markdown
# Logparser Input Plugin

### **Deprecated in 1.10**: Please use the [tail][] plugin along with the
`grok` [data format][].

[tail]: /plugins/inputs/tail/README.md
[data formats]: /docs/DATA_FORMATS_INPUT.md
```

Log a warning message if the plugin is used.  If the plugin is a
ServiceInput, place this in the `Start()` function, for regular Input's log it only the first
time the `Gather` function is called.
```go
log.Println("W! [inputs.logparser] The logparser plugin is deprecated in 1.10. " +
	"Please use the tail plugin with the grok data_format as a replacement.")
```
## Deprecate options

Mark the option as deprecated in the sample config, include the deprecation
version and any replacement.
```toml
  ## Broker URL
  ##   deprecated in 1.7; use the brokers option
  # url = "amqp://localhost:5672/influxdb"
```

In the plugins configuration struct, mention that the option is deprecated:

```go
type AMQPConsumer struct {
	URL string `toml:"url"` // deprecated in 1.7; use brokers
}
```

Finally, use the plugin's `Init() error` method to display a log message at warn level.  The message should include the offending configuration option and any suggested replacement:
```go
func (a *AMQPConsumer) Init() error {
	if p.URL != "" {
		p.Log.Warnf("Use of deprecated configuration: 'url'; please use the 'brokers' option")
	}
	return nil
}
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
