# Request aggregates plugin

The request aggregates plugin generates a set of aggregate values for a response time column in a CSV file within a
given interval. This is especially useful when calculating throughput of systems with high request frequency 
for which storing every single request might require an unnecessary infrastructure. Aggregating values on the client
side minimises the number of writes to the InfluxDB server.

The plugin generates data points at the end of the given window. If no lines were added to the file during a specific
window, no data points are generated.

### Configuration:

```toml
# Aggregates values for requests written to a log file
[[inputs.request_aggregates]]
    # File to monitor.
    file = "/var/server/access.csv"
    # Position of the timestamp of the request in every line
    timestamp_position = 0
    # Format of the timestamp (any layout accepted by Go Time.Parse or s/ms/us/ns for epoch time)
    timestamp_format = "ms"
    # Position of the time value to calculate in the log file (starting from 0)
    time_position = 1
    # Window to consider for time percentiles
    time_window_size = "60s"
    # Windows to keep in memory before flushing in order to avoid requests coming in after a window is shut.
    # If the CSV file is sorted by timestamp, this can be set to 1
    time_windows = 5
    # List of percentiles to calculate
    time_percentiles = [90.0, 95.0, 99.0, 99.99]
    # Position of the result column (success or failure)
    result_position = 3
    # Regular expression used to determine if the result is successful or not (if empty only request_aggregates_all
    # time series) will be generated
    result_success_regex = ".*true.*"
    # Time window to calculate throughput counters
    throughput_window_size = "1s"
    # Number of windows to keep in memory for throughput calculation
    throughput_windows = 300
    # List of tags and their values to add to every data point
    [inputs.aggregates.tags]
    name = "myserver"
```

### Measurements & Fields:
Note: There are as many `perc[_percentile]` as percentiles defined in the configuration.

- request_aggregates
    - requests (integer)
    - time_min (float)
    - time_max (float)
    - time_mean (float)
    - time_perc_90 (float)
    - time_perc_95 (float)
    - [...]
    - time_perc_99_99 (float)
- request_aggregates_success
    - requests (integer)
    - time_min (float)
    - time_max (float)
    - time_mean (float)
    - time_perc_90 (float)
    - time_perc_95 (float)
    - [...]
    - time_perc_99_99 (float)
- request_aggregates_failure
    - requests (integer)
    - time_min (float)
    - time_max (float)
    - time_mean (float)
    - time_perc_90 (float)
    - time_perc_95 (float)
    - [...]
    - time_perc_99_99 (float)
- request_aggregates_throughput
    - requests_total (integer)
    - requests_failed (integer)

### Tags:
Tags are user defined in `[inputs.aggregates.tags]`

### Example output:

```
$ ./telegraf -config telegraf.conf -input-filter request_aggregates -test
request_aggregates,name=myserver requests=186,time_max=380,time_min=86,time_mean=258.54,time_perc_90=200,time_perc_95=220,time_perc_99=225,time_perc_99_99=229 1462270026000000000
request_aggregates_success,name=myserver requests=123,time_max=230,time_min=86,time_mean=120.23,time_perc_90=200,time_perc_95=220,time_perc_99=225,time_perc_99_99=229 1462270026000000000
request_aggregates_failure,name=myserver requests=63,time_max=380,time_min=132,time_mean=298.54,time_perc_90=250,time_perc_95=270,time_perc_99=285,time_perc_99_99=290 1462270026000000000
request_aggregates_throughput,name=myserver requests_total=186,requests_failed=63 1462270026000000000
```