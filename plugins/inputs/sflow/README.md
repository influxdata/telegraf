# SFlow input pluging

The sflow input pluging provides support for acting as an sflow v5 collector in accordance with the specification from [sflow.org](https://sflow.org/)

# Configuration

| Name | Description 
|---|---|
| service_address| URL to listen on expressed as UDP (IPv4 or 6) OP address and port number | service_address = "udp://:6343" 
| dns_multi_name_processor | An optional regexp and template to use to transform a DNS resolve name. Particularily useful when DNS resolves an IP address to more than one name, and they alternative in order when queried. Using this processor command it is possible to tranform the name into something common irrespect of which entry is first - if the names conform to a regular naming schema. Note TOML [escape sequences](https://github.com/toml-lang/toml) may be required.
||For example, ````s/(.*)(?:-net[0-9])/$1```` will strip ```-net<n>``` from the host name thereby converting, as an example, ```hostx-net1``` and ```hostx-net2``` both to ```hostx```
|dns_name_resolve|dfsdf|



```
  ## URL to listen on
  # service_address = "udp://:6343"
  # service_address = "udp4://:6343"
  # service_address = "udp6://:6343"
  
  ## Maximum socket buffer size (in bytes when no unit specified).
  ## Once the buffer fills up, metrics will start dropping.
  ## Defaults to the OS default.
  # read_buffer_size = "64KiB"

  ## The SNMP community string to use for access SNMP on the agents in order to resolve interface names
  # snmp_community = "public"

  ## Whether interface indexes should be turned into interface names via use of snmp
  # snmp_iface_resolve = false

  ## The length of time the interface names are cached
  # snmp_iface_cache_ttl = 3600

  ## Should IP addresses be resolved to host names through DNS lookup
  # dns_name_resolve = false

  ## The length of time the FWDNs are cached
  # dns_name_cache_ttl = 3600

  ##
  # max_flows_per_sample = 10
  # max_counters_per_sample = 10
  # max_samples_per_packet = 10

```

## DNS Name and SNMP Interface name resolution and caching

Raw sflow packets, and their samples headers, communicate IP addresses and Interface identifiers, neither of which are useful to humans.

The sflow plugin can be configured to attempt to resolve IP addresses to host names via DNS and interface identifiers to interface short names via SNMP agent interaction.

The resolved names, or in the case of a resolution error the ip/id will be used as 'the' name, are configurably cached for a period of time to avoid continual lookups.

### Multipe DNS Name resolution & processing

In some cases DNS servers may maintain multiple entries for the same IP address in support of load balancing. In this setup the same IP address may be resolved to multiple DNS names, via a single DNS query, and it is likely the order of those DNS names will change over time.

In order to provide some stability to the names recorded against flow records, it is possible to provide a regular expression and template transformation that should be capable of converting multiple names to a single common name where a mathodical naming scheme has been used.

For example:


# Schema

**ACTUALLY** document this on parser and refere to it

## Tags (optionally as Fields)

The following items are naturally recorded as tags by the sflow plugin. However, using the ```as_fields``` configuration parameter it is possible to have any of these recorded as fields instead

| Name | Description |
|---|---|
| agent_id | IP address of the agent that obtained the sflow sample and sent it to this collector | 
| host | If DNS name resolved is enabled then the host name associated with the agent_id otherwise the agent_id IP address as is|
| source_id_type| NOT DONE (**I don't have this translated**|)
| source_id_index| |
| source_id_name | |
| netif_index_in | |
| netif_name_in | |
| netif_index_out | |
| netif_name_out | |
| sample_direction | |
| header_protocol | |
| ether_type | |
| src_ip | |
| src_host | |
| src_port | |
| src_port_name | |
| src_mac | |
| src_vlan | |
| src_priority | 
| src_mask_len | |
| dst_ip
| dst_host
| dst_port
| dst_port_name
| dst_mac
| dst_vlan
| dst_priority
| dst_mask_len
| next_hop
| ip_version
| ip_protocol
| ip_dscp
| ip_ecn
| tcp_urgent_pointer






## Fields
| Name | Description |
|---|---|
|  bytes |   |
|  drops |   |
|  packets |   |
| frame_length |
| header_size
| ip_fragment_offset
| ip_header_length
| ip_total_length
| ip_ttl
| tcp_header_length
| tcp_window_size
| udp_length
| ip_flags
| tcp_flags | TCP flags of TCP IP header (IPv4 or IPv6)

# TODO

// State config parameters

// State dns resolution control 

// Refere to parser (which has fields)

/// converned about optiopns sizing on header -don't take account of that

// sort example config

// check sflow asrt / break and warn

// test all tags can be turned into fields
// test certain tags are natural type as fields

Resolve this against v0.2 code base
```
2019-09-13T09:55:34Z I! Starting Telegraf sflow~v0.3
2019-09-13T09:55:34Z I! Using config file: /etc/telegraf/telegraf.conf
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x8 pc=0x779a5d]

goroutine 24 [running]:
github.com/influxdata/telegraf/metric.(*metric).GetTag(0xc0025bc8c0, 0x25f16a4, 0xf, 0x0, 0x0, 0x0)
    /Users/deansheehan/go/src/github.com/influxdata/telegraf/metric/metric.go:171 +0x4d
github.com/influxdata/telegraf/plugins/inputs/sflow.(*asyncResolver).ifaceResolve(0xc001824460, 0x2ab1c60, 0xc0025bc8c0, 0x25f16a4, 0xf, 0x25ee6ec, 0xe, 0xc001ad8730, 0xd, 0xc000380d80)
    /Users/deansheehan/go/src/github.com/influxdata/telegraf/plugins/inputs/sflow/resolver.go:167 +0x49
github.com/influxdata/telegraf/plugins/inputs/sflow.(*asyncResolver).resolve(0xc001824460, 0x2ab1c60, 0xc0025bc8c0, 0xc001c90e80)
    /Users/deansheehan/go/src/github.com/influxdata/telegraf/plugins/inputs/sflow/resolver.go:112 +0x2df
github.com/influxdata/telegraf/plugins/inputs/sflow.(*packetSFlowListener).process(0xc0005f8f30, 0xc0001f8000, 0x544, 0x10000)
    /Users/deansheehan/go/src/github.com/influxdata/telegraf/plugins/inputs/sflow/sflow.go:57 +0x220
github.com/influxdata/telegraf/plugins/inputs/sflow.(*packetSFlowListener).listen(0xc0005f8f30)
    /Users/deansheehan/go/src/github.com/influxdata/telegraf/plugins/inputs/sflow/sflow.go:44 +0x75
created by github.com/influxdata/telegraf/plugins/inputs/sflow.(*SFlowListener).Start
    /Users/deansheehan/go/src/github.com/influxdata/telegraf/plugins/inputs/sflow/sflow.go:193 +0x444
```

Also, Timo seems to hae dropped measurement and move more things from tags to fields but is stil lshowing as tags


```
this looks interesting… I dropped the sflow measurement before applying the new config on telegraf. It seems that some of the data is written both as tag and field (dst_mac, src_mac etc)
> show tag keys from sflow
name: sflow
tagKey
------
agent_ip
dst_host
dst_ip
dst_mac
dst_port
dst_port_name
ether_type
header_protocol
host
ip_dscp
ip_ecn
netif_index_in
netif_index_out
netif_name_in
netif_name_out
sample_direction
source_id
source_id_index
source_id_name
src_host
src_ip
src_mac
> show field keys from sflow
name: sflow
fieldKey           fieldType
--------           ---------
bytes              integer
drops              integer
dst_mac            string
frame_length       integer
header_length      integer
ip_flags           integer
ip_fragment_offset integer
ip_total_length    integer
ip_ttl             integer
netif_index_in     string
netif_index_out    string
packets            integer
source_id          string
source_id_index    string
src_mac            string
src_port           string
src_port_name      string
tcp_flags          integer
tcp_header_length  integer
tcp_urgent_pointer integer
tcp_window_size    integer
udp_length         integer
and same mac addresses can be found as tag and field

```



TODO

Put in the SFLow config fields into the config