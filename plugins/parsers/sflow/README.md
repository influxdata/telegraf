# SFlow Parser

## Current Scope

Currently this SFlow Packet parser will only parse SFlow Version 5 and within that only Flow Samples, not Counter Samples.

Within Flow Samples, Ethernet samples with {IPv4, IPv6} and {UDP,TCP} headers are parsed.

# Schema
## Natural Tags
| Name | Description |
|---|---|
| agent_address | IP address of the agent that obtained the sflow sample and sent it to this collector | 
| source_id_type| Decoded from source_id_type field of flow_sample or flow_sample_expanded structures
| source_id_index| Decoded from source_id_index field of flow_sample or flow_sample_expanded structures|
| input_ifindex | Decoded from value (input) field of flow_sample or flow_sample_expanded structures|
| output_ifindex | Decoded from value (output) field of flow_sample or flow_sample_expanded structures|
| sample_direction | Derived from source_id_index, netif_index_in and netif_index_out|
| header_protocol | Decoded from header_protocol field of sampled_header structures|
| ether_type | Decoded from eth_type field of an ETHERNET-ISO88023 header|
| src_ip | Decoded from source_ipaddr field of IPv4 or IPv6 structures|
| src_port | Decoded from src_port field of TCP or UDP structures|
| src_port_name | Derived from src_port|
| src_mac | Decoded from source_mac_addr field of an ETHERNET-ISO88023 header|
| src_vlan | Decoded from src_vlan field of extended_switch structure|
| src_priority | Decoded from src_priority field of extended_switch structure |
| src_mask_len | Decoded from src_mask_len field of extended_router structure|
| dst_ip | Decoded from destination_ipaddr field of IPv4 or IPv6 structures
| dst_port | Decoded from dst_port field of TCP or UDP structures
| dst_port_name | Derived from dst_port
| dst_mac | Decoded from destination_mac_addr field of an ETHERNET-ISO88023 header
| dst_vlan | Decoded from dst_vlan field of extended_switch structure
| dst_priority | Decoded from dst_priority field of extended_switch structure
| dst_mask_len | Decoded from dst_mask_len field of extended_router structure
| next_hop | Decoded from next_hop field of extended_router structure
| ip_version | Decoded from ip_ver field of IPv4 or IPv6 structures
| ip_protocol | Decoded from ip_protocol field of IPv4 or IPv6 structures
| ip_dscp | Decoded from ip_dscp field of IPv4 or IPv6 structures
| ip_ecn | Decoded from ecn field of IPv4 or IPv6 structures
| tcp_urgent_pointer | Decoded from urgent_pointer field of TCP structure

## Fields
| Name | Type | Description |
|---|---|---|
|  bytes |  Integer | Derived from the product of frame_length and packets
|  drops |  Integer |Decoded from drops field of flow_sample or flow_sample_expanded structures
|  packets | Integer |Decoded from sampling_rate field of flow_sample or flow_sample_expanded structures
| frame_length | Integer | Decoded from frame_length field of sampled_header structures
| header_size | Integer | Decoded from header_size field of sampled_header structures
| ip_fragment_offset | Integer | Decoded from ip_ver field of IPv4 structures
| ip_header_length | Integer | Decoded from ip_ver field of IPv4 structures
| ip_total_length | Integer | Decoded from ip_total_len field of IPv4 structures
| ip_ttl | Integer | Decoded from ip_ttl field of IPv4 structures or ip_hop_limit field IPv6 structures
| tcp_header_length | Integer | Decoded from size field of TCP structure. This value is specified in 32-bit words. It must be multiplied by 4 to produce a value in bytes.
| tcp_window_size | Integer | Decoded from window_size field of TCP structure
| udp_length | Integer | 	Decoded from length field of UDP structures
| ip_flags | Integer | Decoded from ip_ver field of IPv4 structures
| tcp_flags | Integer | TCP flags of TCP IP header (IPv4 or IPv6)


