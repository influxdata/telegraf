# Synproxy Input Plugin

The synproxy plugin gathers the synproxy counters. Synproxy is a Linux netfilter
module used for SYN attack mitigation.  The use of synproxy is documented in
`man iptables-extensions` under the SYNPROXY section.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Get synproxy counter statistics from procfs
# This plugin ONLY supports Linux
[[inputs.synproxy]]
  # no configuration
```

The synproxy plugin does not need any configuration

## Metrics

The following synproxy counters are gathered

- synproxy
  - fields:
    - cookie_invalid (uint32, packets, counter) - Invalid cookies
    - cookie_retrans (uint32, packets, counter) - Cookies retransmitted
    - cookie_valid (uint32, packets, counter) - Valid cookies
    - entries (uint32, packets, counter) - Entries
    - syn_received (uint32, packets, counter) - SYN received
    - conn_reopened (uint32, packets, counter) - Connections reopened

## Sample Queries

Get the number of packets per 5 minutes for the measurement in the last hour
from InfluxDB:

```sql
SELECT difference(last("cookie_invalid")) AS "cookie_invalid", difference(last("cookie_retrans")) AS "cookie_retrans", difference(last("cookie_valid")) AS "cookie_valid", difference(last("entries")) AS "entries", difference(last("syn_received")) AS "syn_received", difference(last("conn_reopened")) AS "conn_reopened" FROM synproxy WHERE time > NOW() - 1h GROUP BY time(5m) FILL(null);
```

## Troubleshooting

Execute the following CLI command in Linux to test the synproxy counters:

```sh
cat /proc/net/stat/synproxy
```

## Example Output

This section shows example output in Line Protocol format.

```text
synproxy,host=Filter-GW01,rack=filter-node1 conn_reopened=0i,cookie_invalid=235i,cookie_retrans=0i,cookie_valid=8814i,entries=0i,syn_received=8742i 1549550634000000000
```
