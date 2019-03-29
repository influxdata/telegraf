# BIND 9 Nameserver Statistics Input Plugin

This plugin decodes the JSON or XML statistics provided by BIND 9 nameservers.

### XML Statistics Channel

Version 2 statistics (BIND 9.6 - 9.9) and version 3 statistics (BIND 9.9+) are supported. Note that
for BIND 9.9 to support version 3 statistics, it must be built with the `--enable-newstats` compile
flag, and it must be specifically requested via the correct URL. Version 3 statistics are the
default (and only) XML format in BIND 9.10+.

### JSON Statistics Channel

JSON statistics schema version 1 (BIND 9.10+) is supported. As of writing, some distros still do
not enable support for JSON statistics in their BIND packages.

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
  - name=value (multiple)
- bind_memory
  - total_use
  - in_use
  - block_size
  - context_size
  - lost
- bind_memory_context
  - total
  - in_use

### Tags:

- All measurements
  - url
  - source
  - port
- bind_counter
  - type
  - view (optional)
- bind_memory_context
  - id
  - name

### Sample Queries:

These are some useful queries (to generate dashboards or other) to run against data from this
plugin:

```
SELECT non_negative_derivative(mean(/^A$|^PTR$/), 5m) FROM bind_counter \
WHERE "url" = 'localhost:8053' AND "type" = 'qtype' AND time > now() - 1h \
GROUP BY time(5m), "type"
```

```
name: bind_counter
tags: type=qtype
time                non_negative_derivative_A non_negative_derivative_PTR
----                ------------------------- ---------------------------
1553862000000000000 254.99444444430992        1388.311111111194
1553862300000000000 354                       2135.716666666791
1553862600000000000 316.8666666666977         2130.133333333768
1553862900000000000 309.05000000004657        2126.75
1553863200000000000 315.64999999990687        2128.483333332464
1553863500000000000 308.9166666667443         2132.350000000559
1553863800000000000 302.64999999990687        2131.1833333335817
1553864100000000000 310.85000000009313        2132.449999999255
1553864400000000000 314.3666666666977         2136.216666666791
1553864700000000000 303.2333333331626         2133.8166666673496
1553865000000000000 304.93333333334886        2127.333333333023
1553865300000000000 317.93333333334886        2130.3166666664183
1553865600000000000 280.6666666667443         1807.9071428570896
```
