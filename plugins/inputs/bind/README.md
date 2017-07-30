# BIND 9 Nameserver Statistics Input Plugin

This plugin decodes the JSON or XML statistics provided by BIND 9 nameservers.

### XML Statistics Channel

Version 2 statistics (BIND 9.6 - 9.9) and version 3 statistics (BIND 9.9+) are supported. Note that
for BIND 9.9 to support version 3 statistics, it must be built with the `--enable-newstats` compile
flag, and it must be specifically requested via the correct URL. Version 3 statistics are the
default (and only) XML format in BIND 9.10+.

### JSON Statistics Channel

JSON statistics schema version 1 (BIND 9.10+) is supported. As of writing, most distros do not
currently enable support for JSON statistics in their BIND packages.

### Configuration:

- **urls** []string: List of BIND statistics channel URLs to collect from. Do not include a
  trailing slash in the URL. Default is "http://localhost:8053/xml/v3".
- **gather_memory_contexts** bool: Report per-context memory statistics.
- **gather_views** bool: Report per-view query statistics.

The following table summarizes the URL formats which should be used, depending on your BIND
version and configured statistics channel.

| BIND Version | Statistics Format | Example URL                   |
| ------------ | ----------------- | ----------------------------- |
| 9.6 - 9.8    | XML v2            | http://localhost:8053         |
| 9.9          | XML v2            | http://localhost:8053/xml/v2  |
| 9.9+         | XML v3            | http://localhost:8053/xml/v3  |
| 9.10+        | JSON v1           | http://localhost:8053/json/v1 |

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
