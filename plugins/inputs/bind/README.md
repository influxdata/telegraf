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

- bind_adbstats
- bind_cache
- bind_memory
- bind_memory_context
- bind_opcodes
- bind_nsstats
- bind_resstats
- bind_server
- bind_sockstats
- bind_rdtypes
- bind_zonestats
- bind_view_*

### Tags:

TBC

### Sample Queries:

These are some useful queries (to generate dashboards or other) to run against data from this
plugin:

TBC

### Example Output:

TBC
