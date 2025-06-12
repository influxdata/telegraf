# SFlow Input Plugin

This service plugin produces metrics from information received by acting as a
[SFlow V5][sflow_v5] collector. Currently, the plugin can collect Flow Samples
of Ethernet / IPv4, IPv4 TCP and UDP headers. Counters and other header samples
are ignored. Please use the [netflow plugin][netflow] for a more modern and
sophisticated implementation.

> [!CRITICAL]
> This plugin produces high cardinality data, which when not controlled for will
> cause high load on your database. Please make sure to [filter][filtering] the
> produced metrics or configure your database to avoid cardinality issues!

⭐ Telegraf v1.14.0
🏷️ network
💻 all

[sflow_v5]: https://sflow.org/sflow_version_5.txt
[netflow]: /plugins/inputs/netflow/README.md
[filtering]: /docs/CONFIGURATION.md#metric-filtering

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listen and wait for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# SFlow V5 Protocol Listener
[[inputs.sflow]]
  ## Address to listen for sFlow packets.
  ##   example: service_address = "udp://:6343"
  ##            service_address = "udp4://:6343"
  ##            service_address = "udp6://:6343"
  service_address = "udp://:6343"

  ## Set the size of the operating system's receive buffer.
  ##   example: read_buffer_size = "64KiB"
  # read_buffer_size = ""
```

## Metrics

- sflow
  - tags:
    - agent_address (IP address of the agent that obtained the sflow sample and sent it to this collector)
    - source_id_type(source_id_type field of flow_sample or flow_sample_expanded structures)
    - source_id_index(source_id_index field of flow_sample or flow_sample_expanded structures)
    - input_ifindex (value (input) field of flow_sample or flow_sample_expanded structures)
    - output_ifindex (value (output) field of flow_sample or flow_sample_expanded structures)
    - sample_direction (source_id_index, netif_index_in and netif_index_out)
    - header_protocol (header_protocol field of sampled_header structures)
    - ether_type (eth_type field of an ETHERNET-ISO88023 header)
    - src_ip (source_ipaddr field of IPv4 or IPv6 structures)
    - src_port (src_port field of TCP or UDP structures)
    - src_port_name (src_port)
    - src_mac (source_mac_addr field of an ETHERNET-ISO88023 header)
    - src_vlan (src_vlan field of extended_switch structure)
    - src_priority (src_priority field of extended_switch structure)
    - src_mask_len (src_mask_len field of extended_router structure)
    - dst_ip (destination_ipaddr field of IPv4 or IPv6 structures)
    - dst_port (dst_port field of TCP or UDP structures)
    - dst_port_name (dst_port)
    - dst_mac (destination_mac_addr field of an ETHERNET-ISO88023 header)
    - dst_vlan (dst_vlan field of extended_switch structure)
    - dst_priority (dst_priority field of extended_switch structure)
    - dst_mask_len (dst_mask_len field of extended_router structure)
    - next_hop (next_hop field of extended_router structure)
    - ip_version (ip_ver field of IPv4 or IPv6 structures)
    - ip_protocol (ip_protocol field of IPv4 or IPv6 structures)
    - ip_dscp (ip_dscp field of IPv4 or IPv6 structures)
    - ip_ecn (ecn field of IPv4 or IPv6 structures)
    - tcp_urgent_pointer (urgent_pointer field of TCP structure)
  - fields:
    - bytes (integer, the product of frame_length and packets)
    - drops (integer, drops field of flow_sample or flow_sample_expanded structures)
    - packets (integer, sampling_rate field of flow_sample or flow_sample_expanded structures)
    - frame_length (integer, frame_length field of sampled_header structures)
    - header_size (integer, header_size field of sampled_header structures)
    - ip_fragment_offset (integer, ip_ver field of IPv4 structures)
    - ip_header_length (integer, ip_ver field of IPv4 structures)
    - ip_total_length (integer, ip_total_len field of IPv4 structures)
    - ip_ttl (integer, ip_ttl field of IPv4 structures or ip_hop_limit field IPv6 structures)
    - tcp_header_length (integer, size field of TCP structure. This value is specified in 32-bit words. It must be multiplied by 4 to produce a value in bytes.)
    - tcp_window_size (integer, window_size field of TCP structure)
    - udp_length (integer, length field of UDP structures)
    - ip_flags (integer, ip_ver field of IPv4 structures)
    - tcp_flags (integer, TCP flags of TCP IP header (IPv4 or IPv6))

## Troubleshooting

The [sflowtool][] utility can be used to print sFlow packets, and compared
against the metrics produced by Telegraf.

```sh
sflowtool -p 6343
```

If opening an issue, in addition to the output of sflowtool it will also be
helpful to collect a packet capture.  Adjust the interface, host and port as
needed:

```sh
sudo tcpdump -s 0 -i eth0 -w telegraf-sflow.pcap host 127.0.0.1 and port 6343
```

[sflowtool]: https://github.com/sflow/sflowtool

## Example Output

```text
sflow,agent_address=0.0.0.0,dst_ip=10.0.0.2,dst_mac=ff:ff:ff:ff:ff:ff,dst_port=40042,ether_type=IPv4,header_protocol=ETHERNET-ISO88023,input_ifindex=6,ip_dscp=27,ip_ecn=0,output_ifindex=1073741823,source_id_index=3,source_id_type=0,src_ip=10.0.0.1,src_mac=ff:ff:ff:ff:ff:ff,src_port=443 bytes=1570i,drops=0i,frame_length=157i,header_length=128i,ip_flags=2i,ip_fragment_offset=0i,ip_total_length=139i,ip_ttl=42i,sampling_rate=10i,tcp_header_length=0i,tcp_urgent_pointer=0i,tcp_window_size=14i 1584473704793580447
```
