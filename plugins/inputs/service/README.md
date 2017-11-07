# service Input Plugin

For now this plugin just collects process-specific memory information from `/proc/${PID}/smaps`.

### Configuration

```toml
# Service metrics collector
[[inputs.service]]
  # By default no processes will be included.
  # process_names = ["foo", "bar"]
```