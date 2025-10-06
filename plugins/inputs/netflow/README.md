# Netflow Input Plugin

This service plugin acts as a collector for Netflow v5, Netflow v9 and IPFIX
flow information. The Layer 4 protocol numbers are gathered from the
[official IANA assignments][IANA assignments].
The internal field mappings for Netflow v5 fields are defined according to
[Cisco's Netflow v5 documentation][CISCO NF5], Netflow v9 fields are defined
according to [Cisco's Netflow v9 documentation][CISCO NF9] and the
[ASA extensions][ASA extensions].
Definitions for IPFIX are according to [IANA assignment document][IPFIX doc].

⭐ Telegraf v1.25.0
🏷️ network
💻 all

[IANA assignments]: https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
[CISCO NF5]:        https://www.cisco.com/c/en/us/td/docs/net_mgmt/netflow_collection_engine/3-6/user/guide/format.html#wp1006186
[CISCO NF9]:        https://www.cisco.com/en/US/technologies/tk648/tk362/technologies_white_paper09186a00800a3db9.html
[ASA extensions]:   https://www.cisco.com/c/en/us/td/docs/security/asa/special/netflow/asa_netflow.html
[IPFIX doc]:        https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-nat-type

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
# Netflow v5, Netflow v9 and IPFIX collector
[[inputs.netflow]]
  ## Address to listen for netflow,ipfix or sflow packets.
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
  ##   "ipfix"      -- IPFIX / Netflow v10 protocol (also works for Netflow v9)
  ##   "netflow v5" -- Netflow v5 protocol
  ##   "netflow v9" -- Netflow v9 protocol (also works for IPFIX)
  ##   "sflow v5"   -- sFlow v5 protocol
  # protocol = "ipfix"

  ## Private Enterprise Numbers (PEN) mappings for decoding
  ## This option allows to specify vendor-specific mapping files to use during
  ## decoding.
  # private_enterprise_number_files = []

  ## Log incoming packets for tracing issues
  # log_level = "trace"
```

## Private Enterprise Number mapping

Using the `private_enterprise_number_files` option you can specify mappings for
vendor-specific element-IDs with a PEN specification. The mapping has to be a
comma-separated-file (CSV) containing the element's `ID`, its `name` and the
`data-type`. A comma (`,`) is used as separator and comments are allowed using
the hash (`#`) prefix.
The element `ID` has the form `<pen-number>.<element-id>`, the `name` has to be
a valid field-name and `data-type` denotes the mapping of the raw-byte value to
the field's type. For example

```csv
# PEN.ID, name, data type
35632.349,in_src_osi_sap,hex
35632.471,nprobe_ipv4_address,ip
35632.1028,protocol_ntop,string
35632.1036,l4_srv_port,uint
```

specify four elements (`349`, `471`, `1028` and `1036`) for PEN `35632` (ntop)
with the corresponding name and data-type.

Currently the following `data-type`s are supported:

- `bool`    TruthValue according to [RFC5101][RFC5101]
- `int`     signed integer with 8, 16, 32 or 64 bit
- `uint`    unsigned integer with 8, 16, 32 or 64 bit
- `float32` double-precision floating-point number (32 bit)
- `float64` double-precision floating-point number (64 bit)
- `hex`     hex-encoding of the raw byte sequence with `0x` prefix
- `string`  string interpretation of the raw byte sequence
- `mac`     MAC address
- `ip`      IPv4 or IPv6 address
- `proto`   mapping of layer-4 protocol numbers to names

[RFC5101]: https://www.rfc-editor.org/rfc/rfc5101#section-6.1.5

## Troubleshooting

### `Error template not found` warnings

Those warnings usually occur in cases where Telegraf is restarted or reloaded
while the flow-device is already streaming data.
As background, the Netflow and IPFIX protocols rely on templates sent by the
flow-device to decode fields. Without those templates, it is not clear what the
data-type and size of the payload is and this makes it impossible to correctly
interpret the data. However, templates are sent by the flow-device, usually at
the start of streaming and in regular intervals (configurable in the device) and
Telegraf has no means to trigger sending of the templates. Therefore, we need to
skip the packets until the templates are resent by the device.

## Metrics are missing at the output

The metrics produced by this plugin are not tagged in a connection specific
manner, therefore outputs relying on unique series key (e.g. InfluxDB) require
the metrics to contain tags for the protocol, the connection source and the
connection destination. Otherwise, metrics might be overwritten and are thus
missing.

The required tagging can be achieved using the `converter` processor

```toml
[[processors.converter]]
  [processors.converter.fields]
    tag = ["protocol", "src", "src_port", "dst", "dst_port"]
```

__Please be careful as this will produce metrics with high cardinality!__

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
    - in_packets (uint64, number of incoming packets)
    - tcp_flags (string, TCP flags for the flow)

## Example Output

The specific fields vary for the different protocol versions, here are some
examples

### IPFIX

```text
netflow,source=127.0.0.1,version=IPFIX protocol="tcp",vlan_src=0u,src_tos="0x00",flow_end_ms=1666345513807u,src="192.168.119.100",dst="44.233.90.52",src_port=51008u,total_bytes_exported=0u,flow_end_reason="end of flow",flow_start_ms=1666345513807u,in_total_bytes=52u,in_total_packets=1u,dst_port=443u
netflow,source=127.0.0.1,version=IPFIX src_tos="0x00",src_port=54330u,rev_total_bytes_exported=0u,last_switched=9u,vlan_src=0u,flow_start_ms=1666345513807u,in_total_packets=1u,flow_end_reason="end of flow",flow_end_ms=1666345513816u,in_total_bytes=40u,dst_port=443u,src="192.168.119.100",dst="104.17.240.92",total_bytes_exported=0u,protocol="tcp"
netflow,source=127.0.0.1,version=IPFIX flow_start_ms=1666345513807u,flow_end_ms=1666345513977u,src="192.168.119.100",dst_port=443u,total_bytes_exported=0u,last_switched=170u,src_tos="0x00",in_total_bytes=40u,dst="44.233.90.52",src_port=51024u,protocol="tcp",flow_end_reason="end of flow",in_total_packets=1u,rev_total_bytes_exported=0u,vlan_src=0u
netflow,source=127.0.0.1,version=IPFIX src_port=58246u,total_bytes_exported=1u,flow_start_ms=1666345513806u,flow_end_ms=1666345513806u,in_total_bytes=156u,src="192.168.119.100",rev_total_bytes_exported=0u,last_switched=0u,flow_end_reason="forced end",dst="192.168.119.17",dst_port=53u,protocol="udp",in_total_packets=2u,vlan_src=0u,src_tos="0x00"
netflow,source=127.0.0.1,version=IPFIX protocol="udp",vlan_src=0u,src_port=58879u,dst_port=53u,flow_end_ms=1666345513832u,src_tos="0x00",src="192.168.119.100",total_bytes_exported=1u,rev_total_bytes_exported=0u,flow_end_reason="forced end",last_switched=33u,in_total_bytes=221u,in_total_packets=2u,flow_start_ms=1666345513799u,dst="192.168.119.17"
```

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

### sFlow v5

```text
netflow,source=127.0.0.1,version=sFlowV5 out_errors=0i,out_bytes=3946i,status="up",in_unknown_protocol=4294967295i,out_unicast_packets_total=29i,agent_subid=100000i,interface_type=6i,in_unicast_packets_total=28i,out_dropped_packets=0i,in_bytes=3910i,in_broadcast_packets_total=4294967295i,ip_version="IPv4",agent_ip="192.168.119.184",in_snmp=3i,in_errors=0i,promiscuous=0i,interface=3i,in_mcast_packets_total=4294967295i,in_dropped_packets=0i,sys_uptime=12414i,seq_number=2i,speed=1000000000i,out_mcast_packets_total=4294967295i,out_broadcast_packets_total=4294967295i 12414000000
netflow,source=127.0.0.1,version=sFlowV5 sys_uptime=17214i,agent_ip="192.168.119.184",agent_subid=100000i,seq_number=2i,in_phy_interface=1i,ip_version="IPv4" 17214000000
netflow,source=127.0.0.1,version=sFlowV5 in_errors=0i,out_unicast_packets_total=36i,interface=3i,in_broadcast_packets_total=4294967295i,ip_version="IPv4",speed=1000000000i,out_bytes=4408i,out_mcast_packets_total=4294967295i,status="up",in_snmp=3i,in_mcast_packets_total=4294967295i,out_broadcast_packets_total=4294967295i,promiscuous=0i,in_bytes=5568i,out_dropped_packets=0i,sys_uptime=22014i,agent_subid=100000i,in_unknown_protocol=4294967295i,interface_type=6i,in_dropped_packets=0i,in_unicast_packets_total=37i,out_errors=0i,agent_ip="192.168.119.184",seq_number=3i 22014000000

```
