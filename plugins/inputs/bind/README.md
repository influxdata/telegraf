# BIND 9 Nameserver Statistics Input Plugin

This plugin decodes the XML statistics provided by BIND 9 nameservers. Version 2 statistics
(BIND 9.6+) and version 3 statistics (BIND 9.10+) are supported.

JSON statistics are not currently supported.

### Configuration:

- **urls** []string: List of BIND XML statistics URLs to collect from. Default is
  "http://localhost:8053/".
- **gather_memory_contexts** bool: Report per-context memory statistics.
- **gather_views** bool: Report per-view query statistics.

#### Configuration of BIND Daemon

Add the following to your named.conf if running Telegraf on the same host as the BIND daemon:
```
statistics-channels {
    inet 127.0.0.1 port 8053;
};
```

Alternatively, specify a wildcard address (e.g., 0.0.0.0) or specific IP address of an interface to
configure the BIND daemon to listen on that address. Note that you should secure the statistics
channel with an ACL if it is publicly reachable. Consult the BIND Administrator Reference Manual
for more information.

### Measurements & Fields:

- bind_counter
  - value
- bind_memory
  - TotalUse
  - InUse
  - BlockSize
  - ContextSize
  - Lost
- bind_memory_context
  - Total
  - InUse

### Tags:

- All measurements
  - url
- bind_counter
  - type
  - name
  - view (optional)
- bind_memory_context
  - id
  - name

### Sample Queries:

These are some useful queries (to generate dashboards or other) to run against data from this
plugin:

```
SELECT derivative(mean("value"), 5m) FROM bind_counter \
WHERE "url" = 'localhost:8053' AND "type" = 'qtype' AND time > now() - 1h \
GROUP BY time(5m), "type", "name"
```

### Example Output:

```
name: bind_counter
tags: name=A, type=qtype
time                 derivative
----                 ----------
2017-07-25T16:25:00Z 2.8000000000029104
2017-07-25T16:30:00Z 4.799999999995634
2017-07-25T16:35:00Z 0.4000000000014552
2017-07-25T16:40:00Z 0
2017-07-25T16:45:00Z 1.4666666666671517
2017-07-25T16:50:00Z 2.5333333333328483
2017-07-25T16:55:00Z 61.33333333333576
2017-07-25T17:00:00Z 123.0666666666657
2017-07-25T17:05:00Z 4.5666666666656965
2017-07-25T17:10:00Z 3.0333333333328483

name: bind_counter
tags: name=PTR, type=qtype
time                 derivative
----                 ----------
2017-07-25T16:25:00Z 0
2017-07-25T16:30:00Z 10.033333333333303
2017-07-25T16:35:00Z 72.4666666666667
2017-07-25T16:40:00Z 7.5
2017-07-25T16:45:00Z 0.36666666666678793
2017-07-25T16:50:00Z 0.6333333333332121
2017-07-25T16:55:00Z 0.3333333333334849
2017-07-25T17:00:00Z 0.7666666666664241
2017-07-25T17:05:00Z 1.1666666666669698
2017-07-25T17:10:00Z 0.7333333333331211
```
