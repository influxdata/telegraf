# 128T Metrics Input Plugin

The metrics input plugin collects metrics from a 128T instance.

### Configuration

```toml
# Read metrics from a 128T instance
[[inputs.t128_metrics]]
## Required. The base url for metrics collection
# base_url = "http://localhost:31517/api/v1/router/Fabric128/"

## A socket to use for retrieving metrics - unused by default
# unix_socket = "/var/run/128technology/web-server.sock"

## The maximum number of requests to be in flight at once
# max_simultaneous_requests = 20

## Amount of time allowed to complete a single HTTP request
# timeout = "5s"

## The metrics to collect
# [[inputs.t128_metrics.metric]]
# name = "cpu"
#
# [inputs.t128_metrics.metric.fields]
## Refer to the 128T REST swagger documentation for the list of available metrics
#     key_name = "stats/<path_to_metric>"
#     utilization = "stats/cpu/utilization"
#
## [inputs.t128_metrics.metric.parameters]
#     parameter_name = ["value1", "value2"]
#     core = ["1", "2"]
```
