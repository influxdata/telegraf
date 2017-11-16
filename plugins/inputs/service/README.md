# service Input Plugin

For now this plugin just collects process-specific memory information from `/proc/${PID}/smaps`.

### Configuration

```toml
# Service metrics collector
[[inputs.service_mem]]
  # By default no processes will be included - use exact string used by ps
  # process_names = ["telegraf", "/usr/sbin/nrpe"]
```