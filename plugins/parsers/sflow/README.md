# SFlow Parser

## Current Scope

Currently this SFlow Packet parser will only parse SFlow Version 5 and within that only Flow Samples, not Counter Samples.

Within Flow Samples, Ethernet samples with IPv4 o IPv6 UDP or TCP headers are parsed.

# Schema
## Natural Tags
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
| Name | Type | Description |
|---|---|---|
|  bytes |  Integer |
|  drops |  Integer |
|  packets | Integer |
| frame_length | Integer |
| header_size | Integer |
| ip_fragment_offset | Integer |
| ip_header_length | Integer |
| ip_total_length | Integer |
| ip_ttl | Integer |
| tcp_header_length | Integer
| tcp_window_size | Integer
| udp_length | Integer
| ip_flags | Integer
| tcp_flags | Integer | TCP flags of TCP IP header (IPv4 or IPv6)

# Implementation Approach
This SFlow parser has been developed using a generic packet processing engine making it easy to alter butis not the most efficient in memory or cpu utilisation due to heavy use of map[string]interface for recording generic object tree.

In the future it is expected that a move to a GoLand struture based parser or possible even a zero copy approach is adopted. The primary objective at this stage was to build a parser that we easy to understand and easy to modify prior to optimization.


