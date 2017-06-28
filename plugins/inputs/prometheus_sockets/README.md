# Prometheus Socket Harvester Input Plugin

The prometheus input plugin gathers metrics from UNIX sockets located in directories.

You may instrument your applications with prometheus to bind to an UNIX socket in a directory.
This input plugin will then auto-detect all sockets in this directory and poll them for the prometheus metrics.

This plugin will not work on Windows builds.

### Configuration:

Basic Example

```toml
# Get all metrics from the prometheus socket directory
[[inputs.prometheus_sockets]]
  # An array of directories to harvest sockets from and scrape metrics from.
  socket_paths = ["/var/run/prometheus-sockets", "/tmp/prometheus-sockets"]

  # The URL of the metrics handler from the prometheus client
  # must be identical for all sockets in the directory
  url = "/metrics"
```

### Measurements & Fields & Tags:

The same measurement parsing rules as for the prometheus plugin applies here. In addition each measurement is automatically prefixed with the socket name to avoid collisions.

Example:

```
# served from /var/run/prometheus-sockets/test
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.00010425500000000001
go_gc_duration_seconds{quantile="0.25"} 0.000139108
go_gc_duration_seconds{quantile="0.5"} 0.00015749400000000002
go_gc_duration_seconds{quantile="0.75"} 0.000331463
go_gc_duration_seconds{quantile="1"} 0.000667154
go_gc_duration_seconds_sum 0.0018183950000000002
go_gc_duration_seconds_count 7
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 15
```

This will generate the following measurements
- test_go_gc_duration_seconds
- test_go_goroutines
