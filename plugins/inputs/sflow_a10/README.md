# SFlowA10 Input Plugin

The SFlow_A10 Input Plugin provides support for acting as an SFlow V5 collector
for [A10](https://www.a10networks.com/) appliances, in accordance with the
specification from [sflow.org](https://sflow.org/).

It is heavily based (i.e. re-uses a lot of code and techniques) on the
[SFlow](../sflow/README.md) plugin. The main difference is that SFlow_A10
 captures ony Counter Samples that are coming from A10 appliance and turns them
 into telegraf metrics. Flow samples and header samples are ignored.

## How this works

Plugin starts by reading the XML file with the counter record definitions.
Counter records which definition is not included in the XML file are ignored
when they arrive at the plugin.

The way that the plugin works is that it parses incoming counter records from
A10. When it discovers counter records tagged 260 (port information) or 271/272
(IPv4/IPv6 information) it parses their sourceID and stores them in memory.
When a counter record metric arrives, plugin checks if there is port and ip
information for it (i.e. we have gotten 260 and 271 or 272 for the same
sourceID). If there is, the metric is sent to telegraf output.
If it is not, the metric is discarded.

## Configuration

```toml
[[inputs.sflow_a10]]
  ## Address to listen for sFlow packets.
  ##   example: service_address = "udp://:6343"
  ##            service_address = "udp4://:6343"
  ##            service_address = "udp6://:6343"
  service_address = "udp://:6343"

  ## Set the size of the operating system's receive buffer.
  ##   example: read_buffer_size = "64KiB"
  # read_buffer_size = ""

  # XML file containing counter definitions, according to A10 specification
  a10_definitions_file = "/path/to/xml_file.xml"

  # if true, metrics with zero values will not be sent to the output
  # this is to lighten the load on the metrics database backend
  ignore_zero_values = true
```

## Metrics

- sflow_a10
  - tags:
    - agent_address
    - ip_address
    - port_number
    - port_range_end
    - port_type
    - table_ type
  - fields:
    - all counters that are included in the XML file

## Example Output

```shell
sflow_a10,agent_address=10.1.0.6,ip_address=10.3.0.39,port_number=0,port_range_end=0,port_type=INVALID,table_type=Zone udp_total_bytes_forwarded_diff=0,src_dst_pair_entry_total_count_diff=0,inbound_packets_dropped_diff=25,tcp_total_bytes_dropped_diff=1932,udp_total_bytes_dropped_diff=0,tcp_connections_created_from_syn_diff=3,tcp_connections_closed_diff=10,outbound_bytes_forwarded_diff=1776,udp_dst_port_total_exceeded_diff=0,src_dst_pair_entry_tcp_count_diff=0,tcp_connections_created_from_ack_diff=6,tcp_total_bytes_received_diff=2418,sflow_external_samples_packed_diff=1,sflow_external_packets_sent_diff=3,inbound_bytes_dropped_diff=1932,udp_total_bytes_received_diff=0,tcp_total_bytes_forwarded_diff=2262 1605280935424500515
```
