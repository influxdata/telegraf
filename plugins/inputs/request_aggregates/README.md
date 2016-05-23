# Request aggregates plugin

Measuring request throughput and elapsed time of web/application servers is important both in production systems and to
evaluate load testing performance. Most servers (e.g. NGINX, Tomcat) and load runners (e.g. JMeter, SoapUI) enable users
to configure the format of their access logs, making CSV a widely used format for future analysis.
The Request Aggregates plugin monitors a log file in CSV format and generates a set of aggregated data points for
throughput and elapsed time within certain given time windows. This is particularly useful when calculating throughput
of systems with high request frequency (thousands per second) for which storing every single request might require an
unnecessary infrastructure. Aggregating values on the client side minimises the number of writes to the configured
outputs.

The following is an example of a CSV file to use with this plugin (as mentioned below, the positions and formats of
key fields are configurable):

```
1456761335505,1117,HTTP Request,200,OK,"Everything is working",text,true
1456761335307,1385,HTTP Request,200,OK,"Everything is working, still",text,true
1456761335200,1494,HTTP Request,200,OK,"I'm noticing some strange behaviour, this request is not successful",text,false
1456761335265,1431,HTTP Request,500,Internal Server Error,"Oh no! Something strange happened!",text,false
1456761335211,1491,HTTP Request,200,OK,"Everything is working again, phew!",text,true
```

The plugin tails a given file and puts requests corresponding to new lines in their respective time windows. As requests
do not necessarily need to be sorted by time in the log file, the number of windows to be kept in memory can be
configured. When a window expires (if we have 5 windows of 1 minute that would be in 5 minutes) the values in that
window are aggregated and metrics are flushed. After a window has been aggregated and flushed, any new request belonging
to that window will be omitted (and logged as an error in the Telegraf log).

### Configuration:

```toml
# Aggregates values for requests written to a CSV file
[[inputs.request_aggregates]]
    # File to monitor (must be in CSV format).
    file = "/var/server/access.csv"
    # Position of the timestamp of the request in every line.
    timestamp_position = 0
    # Format of the timestamp (any layout accepted by Go Time.Parse or s/ms/us/ns for epoch time).
    timestamp_format = "ms"
    # Position, starting from 0, of the elapsed time value to calculate in the log file.
    time_position = 1
    # Window to consider for elapsed time percentiles.
    time_window_size = "60s"
    # Windows to keep in memory before flushing.
    time_windows = 5
    # List of percentiles to calculate (must be float numbers)
    time_percentiles = [90.0, 95.0, 99.0, 99.99]
    # Position of the result column to determine if a request is successful or not.
    result_position = 7
    # Regular expression used to determine if the result is successful or not (if empty only request_aggregates_total
    # measurement) will be generated
    result_success_regex = ".*true.*"
    # Time window to calculate throughput counters
    throughput_window_size = "1s"
    # Number of windows to keep in memory for throughput calculation
    throughput_windows = 300
    # List of tags and their values to add to every data point
    [inputs.request_aggregates.tags]
    server_name = "myserver"
```

### Measurements & Fields:
Note: There are as many `perc[_percentile]` as percentiles defined in the configuration.

- request_aggregates_total
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
Tags are user defined in `[inputs.request_aggregates.tags]`

### Example output:

```
request_aggregates_total,server_name=myserver requests=186,time_max=380,time_min=86,time_mean=258.54,time_perc_90=200,time_perc_95=220,time_perc_99=225,time_perc_99_99=229 1462270026000000000
request_aggregates_success,server_name=myserver requests=123,time_max=230,time_min=86,time_mean=120.23,time_perc_90=200,time_perc_95=220,time_perc_99=225,time_perc_99_99=229 1462270026000000000
request_aggregates_failure,server_name=myserver requests=63,time_max=380,time_min=132,time_mean=298.54,time_perc_90=250,time_perc_95=270,time_perc_99=285,time_perc_99_99=290 1462270026000000000
request_aggregates_throughput,server_name=myserver requests_total=186,requests_failed=63 1462270026000000000
```