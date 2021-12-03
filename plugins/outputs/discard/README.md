# discard Output Plugin

This output plugin simply drops all metrics that are sent to it. It is only
meant to be used for testing purposes.

## Configuration

```toml
# Send metrics to nowhere at all
[[outputs.discard]]
  # no configuration
```
