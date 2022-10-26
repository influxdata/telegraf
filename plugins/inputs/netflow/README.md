# Netflow Input Plugin

The `netflow` plugin acts as a collector for Netflow (v9) and IPFIX
flow information. The Layer 4 protocol numbers are gathered from the
[official IANA assignments][IANA assignments].
The internal field mappings for the Netflow v9 fields are defined according to
[Cisco's Netflow9 documentation][CISCO NF9] and the [ASA extensions][].
Definitions for IPFIX are according to [IANA assignement document][IPFIX doc].

[IANA assignments]: https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
[CISCO NF9]:        https://www.cisco.com/en/US/technologies/tk648/tk362/technologies_white_paper09186a00800a3db9.html
[ASA extensions]:   https://www.cisco.com/c/en/us/td/docs/security/asa/special/netflow/asa_netflow.html
[IPFIX doc]:        https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-nat-type
## Configuration

```toml @sample.conf
# Netflow V9 and IPFIX collector
[[inputs.netflow]]
  ## Address to listen for netflow/ipfix packets.
  ##   example: service_address = "udp://:2055"
  ##            service_address = "udp4://:2055"
  ##            service_address = "udp6://:2055"
  service_address = "udp://:2055"

  ## Set the size of the operating system's receive buffer.
  ##   example: read_buffer_size = "64KiB"
  # read_buffer_size = ""

  ## Dump incoming packets to the log
  ## This can be helpful to debug parsing issues. Only active if
  ## Telegraf is in debug mode.
  # dump_packets = false
```

## Metrics

Metrics depend on the format used as well as on the information provided
by the exporter. Furthermore, proprietary information might be sent requiring
further decoding information. Most exporters should provide at least the
following information

- netflow
  - tags:
    - source (IP of the exporter sending the data)
    - version (flow protocol version)
  - fields:
    - src (IP address, address of the source of the packets)
    - dst (IP address, address of the destination of the packets)
    - l4_src_port (uint64, source port)
    - l4_dst_port (uint64, destination port)
    - l4_protocol (string, Layer 4 protocol name)
    - in_bytes (uint64, number of incoming bytes)
    - in_packets (uint64, number of incomming packets)
    - flow_create_time_ms (optional, time the flow was created)
    - flow_end_time_ms (optional, time the flow ended)

## Sample Queries

This section can contain some useful InfluxDB queries that can be used to get
started with the plugin or to generate dashboards.  For each query listed,
describe at a high level what data is returned.

Get the max, mean, and min for the measurement in the last hour:

```sql
SELECT max(field1), mean(field1), min(field1) FROM measurement1 WHERE tag1=bar AND time > now() - 1h GROUP BY tag
```

## Troubleshooting

This optional section can provide basic troubleshooting steps that a user can
perform.

## Example

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter <plugin-name> --test` or use the `file` output to get
this information.

```shell
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```
