# SFlow input pluging

The sflow input pluging provides support for acting as an sflow v5 collector in accordance with the specification from [sflow.org](https://sflow.org/)

# Configuration

| Name | Description | Example |
|---|---|---|
| service_address| URL to listen on expressed as UDP (IPv4 or 6) OP address and port number | service_address = "udp://:6343" 
| dns_multi_name_processor | An optional regexp and template to use to transform a DNS resolve name. Particularily useful when DNS resolves an IP address to more than one name, and they alternative in order when queried. Using this processor command it is possible to tranform the name into something common irrespect of which entry is first - if the names conform to a regular naming schema. Note TOML [escape sequences](https://github.com/toml-lang/toml) may be required. | ````s/(.*)(?:-net[0-9])/$1```` will strip ```-net<n>``` from the host name thereby converting, as an example, ```hostx-net1``` and ```hostx-net2``` both to ```hostx```



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

# Schema

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
|   |   |
|   |   |
|   |   |

# TODO

// State config parameters

// State dns resolution control 

// Refere to parser (which has fields)

// find sflow sample headers to test against

// find ehternet udp/tcp headers to test against

// converned about optiopns sizing on header -don't take account of that

// create good tests

