# Netflow Input Plugin

The `netflow` plugin acts as a collector for Netflow v5, Netflow v9 and IPFIX
flow information. The Layer 4 protocol numbers are gathered from the
[official IANA assignments][IANA assignments].
The internal field mappings for Netflow v5 fields are defined according to
[Cisco's Netflow v5 documentation][CISCO NF5], Netflow v9 fields are defined
according to [Cisco's Netflow v9 documentation][CISCO NF9] and the
[ASA extensions][ASA extensions].
Definitions for IPFIX are according to [IANA assignement document][IPFIX doc].

[IANA assignments]: https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
[CISCO NF5]:        https://www.cisco.com/c/en/us/td/docs/net_mgmt/netflow_collection_engine/3-6/user/guide/format.html#wp1006186
[CISCO NF9]:        https://www.cisco.com/en/US/technologies/tk648/tk362/technologies_white_paper09186a00800a3db9.html
[ASA extensions]:   https://www.cisco.com/c/en/us/td/docs/security/asa/special/netflow/asa_netflow.html
[IPFIX doc]:        https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-nat-type

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
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
# Netflow v5, Netflow v9 and IPFIX collector
[[inputs.netflow]]
  ## Address to listen for netflow/ipfix packets.
  ##   example: service_address = "udp://:2055"
  ##            service_address = "udp4://:2055"
  ##            service_address = "udp6://:2055"
  service_address = "udp://:2055"

  ## Set the size of the operating system's receive buffer.
  ##   example: read_buffer_size = "64KiB"
  ## Uses the system's default if not set.
  # read_buffer_size = ""

  ## Protocol version to use for decoding.
  ## Available options are
  ##   "netflow v5" -- Netflow v5 protocol
  ##   "netflow v9" -- Netflow v9 protocol (also works for IPFIX)
  ##   "ipfix"      -- IPFIX / Netflow v10 protocol (also works for Netflow v9)
  # protocol = "ipfix"

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
    - src_mask (uint64, mask for the IP address in bits)
    - dst (IP address, address of the destination of the packets)
    - dst_mask (uint64, mask for the IP address in bits)
    - src_port (uint64, source port)
    - dst_port (uint64, destination port)
    - protocol (string, Layer 4 protocol name)
    - in_bytes (uint64, number of incoming bytes)
    - in_packets (uint64, number of incomming packets)
    - tcp_flags (string, TCP flags for the flow)

## Example Output

The specific fields vary for the different protocol versions, here are some
examples

### Netflow v5

```text
netflow,source=127.0.0.1,version=NetFlowV5 protocol="tcp",src="140.82.121.3",src_port=443u,dst="192.168.119.100",dst_port=55516u,flows=8u,in_bytes=87477u,in_packets=78u,first_switched=86400660u,last_switched=86403316u,tcp_flags="...PA...",engine_type="19",engine_id="0x56",sys_uptime=90003000u,src_tos="0x00",bgp_src_as=0u,bgp_dst_as=0u,src_mask=0u,dst_mask=0u,in_snmp=0u,out_snmp=0u,next_hop="0.0.0.0",seq_number=0u,sampling_interval=0u
netflow,source=127.0.0.1,version=NetFlowV5 protocol="tcp",src="140.82.121.6",src_port=443u,dst="192.168.119.100",dst_port=36408u,flows=8u,in_bytes=5009u,in_packets=21u,first_switched=86400447u,last_switched=86403267u,tcp_flags="...PA...",engine_type="19",engine_id="0x56",sys_uptime=90003000u,src_tos="0x00",bgp_src_as=0u,bgp_dst_as=0u,src_mask=0u,dst_mask=0u,in_snmp=0u,out_snmp=0u,next_hop="0.0.0.0",seq_number=0u,sampling_interval=0u
netflow,source=127.0.0.1,version=NetFlowV5 protocol="tcp",src="140.82.112.22",src_port=443u,dst="192.168.119.100",dst_port=39638u,flows=8u,in_bytes=925u,in_packets=6u,first_switched=86400324u,last_switched=86403214u,tcp_flags="...PA...",engine_type="19",engine_id="0x56",sys_uptime=90003000u,src_tos="0x00",bgp_src_as=0u,bgp_dst_as=0u,src_mask=0u,dst_mask=0u,in_snmp=0u,out_snmp=0u,next_hop="0.0.0.0",seq_number=0u,sampling_interval=0u
netflow,source=127.0.0.1,version=NetFlowV5 protocol="tcp",src="140.82.114.26",src_port=443u,dst="192.168.119.100",dst_port=49398u,flows=8u,in_bytes=250u,in_packets=2u,first_switched=86403131u,last_switched=86403362u,tcp_flags="...PA...",engine_type="19",engine_id="0x56",sys_uptime=90003000u,src_tos="0x00",bgp_src_as=0u,bgp_dst_as=0u,src_mask=0u,dst_mask=0u,in_snmp=0u,out_snmp=0u,next_hop="0.0.0.0",seq_number=0u,sampling_interval=0u
netflow,source=127.0.0.1,version=NetFlowV5 protocol="tcp",src="192.168.119.100",src_port=55516u,dst="140.82.121.3",dst_port=443u,flows=8u,in_bytes=4969u,in_packets=37u,first_switched=86400652u,last_switched=86403269u,tcp_flags="...PA...",engine_type="19",engine_id="0x56",sys_uptime=90003000u,src_tos="0x00",bgp_src_as=0u,bgp_dst_as=0u,src_mask=0u,dst_mask=0u,in_snmp=0u,out_snmp=0u,next_hop="0.0.0.0",seq_number=0u,sampling_interval=0u
netflow,source=127.0.0.1,version=NetFlowV5 protocol="tcp",src="192.168.119.100",src_port=36408u,dst="140.82.121.6",dst_port=443u,flows=8u,in_bytes=2736u,in_packets=21u,first_switched=86400438u,last_switched=86403258u,tcp_flags="...PA...",engine_type="19",engine_id="0x56",sys_uptime=90003000u,src_tos="0x00",bgp_src_as=0u,bgp_dst_as=0u,src_mask=0u,dst_mask=0u,in_snmp=0u,out_snmp=0u,next_hop="0.0.0.0",seq_number=0u,sampling_interval=0u
netflow,source=127.0.0.1,version=NetFlowV5 protocol="tcp",src="192.168.119.100",src_port=39638u,dst="140.82.112.22",dst_port=443u,flows=8u,in_bytes=1560u,in_packets=6u,first_switched=86400225u,last_switched=86403255u,tcp_flags="...PA...",engine_type="19",engine_id="0x56",sys_uptime=90003000u,src_tos="0x00",bgp_src_as=0u,bgp_dst_as=0u,src_mask=0u,dst_mask=0u,in_snmp=0u,out_snmp=0u,next_hop="0.0.0.0",seq_number=0u,sampling_interval=0u
netflow,source=127.0.0.1,version=NetFlowV5 protocol="tcp",src="192.168.119.100",src_port=49398u,dst="140.82.114.26",dst_port=443u,flows=8u,in_bytes=697u,in_packets=4u,first_switched=86403030u,last_switched=86403362u,tcp_flags="...PA...",engine_type="19",engine_id="0x56",sys_uptime=90003000u,src_tos="0x00",bgp_src_as=0u,bgp_dst_as=0u,src_mask=0u,dst_mask=0u,in_snmp=0u,out_snmp=0u,next_hop="0.0.0.0",seq_number=0u,sampling_interval=0u
```

### Netflow v9

```text
netflow,source=127.0.0.1,version=NetFlowV9 protocol="tcp",src="140.82.121.3",src_port=443u,dst="192.168.119.100",dst_port=55516u,in_bytes=87477u,in_packets=78u,flow_start_ms=1666350478660u,flow_end_ms=1666350481316u,tcp_flags="...PA...",engine_type="17",engine_id="0x01",icmp_type=0u,icmp_code=0u,fwd_status="unknown",fwd_reason="unknown",src_tos="0x00"
netflow,source=127.0.0.1,version=NetFlowV9 protocol="tcp",src="140.82.121.6",src_port=443u,dst="192.168.119.100",dst_port=36408u,in_bytes=5009u,in_packets=21u,flow_start_ms=1666350478447u,flow_end_ms=1666350481267u,tcp_flags="...PA...",engine_type="17",engine_id="0x01",icmp_type=0u,icmp_code=0u,fwd_status="unknown",fwd_reason="unknown",src_tos="0x00"
netflow,source=127.0.0.1,version=NetFlowV9 protocol="tcp",src="140.82.112.22",src_port=443u,dst="192.168.119.100",dst_port=39638u,in_bytes=925u,in_packets=6u,flow_start_ms=1666350478324u,flow_end_ms=1666350481214u,tcp_flags="...PA...",engine_type="17",engine_id="0x01",icmp_type=0u,icmp_code=0u,fwd_status="unknown",fwd_reason="unknown",src_tos="0x00"
netflow,source=127.0.0.1,version=NetFlowV9 protocol="tcp",src="140.82.114.26",src_port=443u,dst="192.168.119.100",dst_port=49398u,in_bytes=250u,in_packets=2u,flow_start_ms=1666350481131u,flow_end_ms=1666350481362u,tcp_flags="...PA...",engine_type="17",engine_id="0x01",icmp_type=0u,icmp_code=0u,fwd_status="unknown",fwd_reason="unknown",src_tos="0x00"
netflow,source=127.0.0.1,version=NetFlowV9 protocol="tcp",src="192.168.119.100",src_port=55516u,dst="140.82.121.3",dst_port=443u,in_bytes=4969u,in_packets=37u,flow_start_ms=1666350478652u,flow_end_ms=1666350481269u,tcp_flags="...PA...",engine_type="17",engine_id="0x01",icmp_type=0u,icmp_code=0u,fwd_status="unknown",fwd_reason="unknown",src_tos="0x00"
netflow,source=127.0.0.1,version=NetFlowV9 protocol="tcp",src="192.168.119.100",src_port=36408u,dst="140.82.121.6",dst_port=443u,in_bytes=2736u,in_packets=21u,flow_start_ms=1666350478438u,flow_end_ms=1666350481258u,tcp_flags="...PA...",engine_type="17",engine_id="0x01",icmp_type=0u,icmp_code=0u,fwd_status="unknown",fwd_reason="unknown",src_tos="0x00"
netflow,source=127.0.0.1,version=NetFlowV9 protocol="tcp",src="192.168.119.100",src_port=39638u,dst="140.82.112.22",dst_port=443u,in_bytes=1560u,in_packets=6u,flow_start_ms=1666350478225u,flow_end_ms=1666350481255u,tcp_flags="...PA...",engine_type="17",engine_id="0x01",icmp_type=0u,icmp_code=0u,fwd_status="unknown",fwd_reason="unknown",src_tos="0x00"
netflow,source=127.0.0.1,version=NetFlowV9 protocol="tcp",src="192.168.119.100",src_port=49398u,dst="140.82.114.26",dst_port=443u,in_bytes=697u,in_packets=4u,flow_start_ms=1666350481030u,flow_end_ms=1666350481362u,tcp_flags="...PA...",engine_type="17",engine_id="0x01",icmp_type=0u,icmp_code=0u,fwd_status="unknown",fwd_reason="unknown",src_tos="0x00"
```

### IPFIX

```text
netflow,source=127.0.0.1,version=IPFIX protocol="tcp",vlan_src=0u,src_tos="0x00",flow_end_ms=1666345513807u,src="192.168.119.100",dst="44.233.90.52",src_port=51008u,total_bytes_exported=0u,flow_end_reason="end of flow",flow_start_ms=1666345513807u,in_total_bytes=52u,in_total_packets=1u,dst_port=443u
netflow,source=127.0.0.1,version=IPFIX src_tos="0x00",src_port=54330u,rev_total_bytes_exported=0u,last_switched=9u,vlan_src=0u,flow_start_ms=1666345513807u,in_total_packets=1u,flow_end_reason="end of flow",flow_end_ms=1666345513816u,in_total_bytes=40u,dst_port=443u,src="192.168.119.100",dst="104.17.240.92",total_bytes_exported=0u,protocol="tcp"
netflow,source=127.0.0.1,version=IPFIX flow_start_ms=1666345513807u,flow_end_ms=1666345513977u,src="192.168.119.100",dst_port=443u,total_bytes_exported=0u,last_switched=170u,src_tos="0x00",in_total_bytes=40u,dst="44.233.90.52",src_port=51024u,protocol="tcp",flow_end_reason="end of flow",in_total_packets=1u,rev_total_bytes_exported=0u,vlan_src=0u
netflow,source=127.0.0.1,version=IPFIX src_port=58246u,total_bytes_exported=1u,flow_start_ms=1666345513806u,flow_end_ms=1666345513806u,in_total_bytes=156u,src="192.168.119.100",rev_total_bytes_exported=0u,last_switched=0u,flow_end_reason="forced end",dst="192.168.119.17",dst_port=53u,protocol="udp",in_total_packets=2u,vlan_src=0u,src_tos="0x00"
netflow,source=127.0.0.1,version=IPFIX protocol="udp",vlan_src=0u,src_port=58879u,dst_port=53u,flow_end_ms=1666345513832u,src_tos="0x00",src="192.168.119.100",total_bytes_exported=1u,rev_total_bytes_exported=0u,flow_end_reason="forced end",last_switched=33u,in_total_bytes=221u,in_total_packets=2u,flow_start_ms=1666345513799u,dst="192.168.119.17"
```
