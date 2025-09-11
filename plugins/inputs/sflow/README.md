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
    - agent_address      - IP address of the agent obtaining the sflow sample and
                           sent it to this collector
    - source_id_type     - type of flow_sample or flow_sample_expanded structures
    - source_id_index    - index of flow_sample or flow_sample_expanded structures
    - input_ifindex      - input value of flow_sample or flow_sample_expanded structures
    - output_ifindex     - output value of flow_sample or flow_sample_expanded structures
    - sample_direction   - source_id_index, netif_index_in and netif_index_out
    - header_protocol    - header_protocol field of sampled_header structures
    - ether_type         - ethernet type of an ETHERNET-ISO88023 header
    - src_ip             - source IP address of IPv4 or IPv6 structures
    - src_port           - source port of TCP or UDP structures
    - src_port_name      - name of the source port
    - src_mac            - source MAC address of an ETHERNET-ISO88023 header
    - src_vlan           - source VLAN of extended_switch structure
    - src_priority       - source priority of extended_switch structure
    - src_mask_len       - length of source mask of extended_router structure
    - dst_ip             - destination IP address of IPv4 or IPv6 structures
    - dst_port           - destination port of TCP or UDP structures
    - dst_port_name      - name of the destination port
    - dst_mac            - destination MAC address of an ETHERNET-ISO88023 header
    - dst_vlan           - destination VLAN of extended_switch structure
    - dst_priority       - destination priority extended_switch structure
    - dst_mask_len       - length of destinationd mask of extended_router structure
    - next_hop           - next hop of extended_router structure
    - ip_version         - IP version of IPv4 or IPv6 structures
    - ip_protocol        - IP protocol of IPv4 or IPv6 structures
    - ip_dscp            - IP DSCP of IPv4 or IPv6 structures
    - ip_ecn             - IP ECN of IPv4 or IPv6 structures
    - tcp_urgent_pointer - urgent pointer of TCP structure
  - fields:
    - bytes              (int) - product of frame length and packets
    - drops              (int) - drops field of flow_sample or
                                 flow_sample_expanded structures
    - packets            (int) - sampling_rate field of flow_sample or
                                 flow_sample_expanded structures
    - frame_length       (int) - frame_length field of sampled_header structures
    - header_size        (int) - header_size field of sampled_header structures
    - ip_fragment_offset (int) - ip_ver field of IPv4 structures
    - ip_header_length   (int) - ip_ver field of IPv4 structures
    - ip_total_length    (int) - ip_total_len field of IPv4 structures
    - ip_ttl             (int) - ip_ttl field of IPv4 structures or
                                 ip_hop_limit field IPv6 structures
    - tcp_header_length  (int) - size field of TCP structure. This value is
                                 specified in 32-bit words. It must be multiplied
                                 by 4 to produce a valuein bytes.
    - tcp_window_size    (int) - window_size field of TCP structure
    - udp_length         (int) - length field of UDP structures
    - ip_flags           (int) - ip_ver field of IPv4 structures
    - tcp_flags          (int) - TCP flags of TCP IP header (IPv4 or IPv6)

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
