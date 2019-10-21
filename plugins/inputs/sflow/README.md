# SFlow Input Plugin

The SFlow Input Plugin provides support for acting as an SFlow V5 collector in accordance with the specification from [sflow.org](https://sflow.org/).

Currently only Flow Samples of Ethernet / IPv4 & IPv4 TCP & UDP headers are turned into metrics - counters and other header samples may come later.

# Configuration
The following configuration options are availabe:

| Name | Description 
|---|---|
| service_address| URL to listen on expressed as UDP (IPv4 or 6) OP address and port number 
| | Example: ```service_address = "udp://:6343"```
| read_buffer_size | Maximum socket buffer size (in bytes when no unit specified). Once the buffer fills up, metrics will start dropping. Defaults to the OS default.
||Example = ```read_buffer_size"64KiB"``` |
| dns_multi_name_processor | An optional regexp and template to use to transform a DNS resolve name. Particularily useful when DNS resolves an IP address to more than one name, and they alternative in order when queried. Using this processor command it is possible to tranform the name into something common irrespect of which entry is first - if the names conform to a regular naming schema. Note TOML [escape sequences](https://github.com/toml-lang/toml) may be required.
||Example: ````s/(.*)(?:-net[0-9])/$1```` will strip ```-net<n>``` from the host name thereby converting, as an example, ```hostx-net1``` and ```hostx-net2``` both to ```hostx```
|dns_fqdn_resolve|Determines whether IP addresses should be resolved to Host names.
||Example: ```dns_fqdn_resolve = true```
|dns_fqdn_cache_ttl|The time to live for entries in the DNS name cache expressed in seconds. Default is 0 which is infinite
||Example: ```dns_fwdn_cache_ttl = 3600```
|snmp_iface_resolve = false|Determines whether interface indexes should be looked up using SNMP to provide the natural show name|
||Example: ```snmp_iface_resolve = true```
|snmp_community|The SNMP community string to use for access SNMP on the agents in order to resolve interface names
||Example: ```snmp_community = "public"```
|snmp_iface_cache_ttl| The time to live for entries in the SNMP Interface cache expressed in seconds. Default is 0 which is infinite.
||Example: ```snmp_iface_cache_ttl = 3600```

## Configuration:

This is a sample configuration for the plugin.

```toml
[[inputs.sflow]]
	## URL to listen on
	# service_address = "udp://:6343"
	# service_address = "udp4://:6343"
	# service_address = "udp6://:6343"
    
	## Maximum socket buffer size (in bytes when no unit specified).
	## For stream sockets, once the buffer fills up, the sender will start backing up.
	## For datagram sockets, once the buffer fills up, metrics will start dropping.
	## Defaults to the OS default.
	# read_buffer_size = "64KiB"

	# Whether IP addresses should be resolved to host names
	# dns_fqdn_resolve = true

	# How long should resolved IP->Hostnames be cached (in seconds)
	# dns_fqdn_cache_ttl = 3600
	
	# Optional processing instructions for transforming DNS resolve host names
	# dns_multi_name_processor = "s/(.*)(?:-net[0-9])/$1"

	# Whether Interface Indexes should be resolved to Interface Names via SNMP
	# snmp_iface_resolve = true
	
	# SNMP Community string to use when resolving Interface Names
	# snmp_community = "public"

	# How long should resolved Iface Index->Iface Name be cached (in seconds)
	# snmp_iface_cache_ttl = 3600
```

## DNS Name and SNMP Interface name resolution and caching

Raw SFlow packets, and their samples headers, communicate IP addresses and Interface identifiers, neither of which are useful to humans.

The sflow plugin can be configured to attempt to resolve IP addresses to host names via DNS and interface identifiers to interface short names via SNMP agent interaction.

The resolved names, or in the case of a resolution error the ip/id will be used as 'the' name, are configurably cached for a period of time to avoid continual lookups.

| Source IP Tag | Resolved Host Tag 
|---|---|
|agent_address|agent_host
|src_ip|src_host
|dst_ip|dst_host

| Source IFace Index Tag | Resolved IFace Name Tag 
|---|---|
|source_id_index|source_id_name
|input_ifindex|input_ifname
|output_ifindex|output_ifname

### Multipe DNS Name resolution & processing

In some cases DNS servers may maintain multiple entries for the same IP address in support of load balancing. In this setup the same IP address may be resolved to multiple DNS names, via a single DNS query, and it is likely the order of those DNS names will change over time.

In order to provide some stability to the names recorded against flow records, it is possible to provide a regular expression and template transformation that should be capable of converting multiple names to a single common name where a mathodical naming scheme has been used.

Example: ````s/(.*)(?:-net[0-9])/$1```` will strip ```-net<n>``` from the host name thereby converting, as an example, ```hostx-net1``` and ```hostx-net2``` both to ```hostx```

# Schema

The parsing of SFlow packets is handled by the SFlow Parser and the schema is described [here](../../parsers/sflow/README.md).

At a high level, individual Flow Samples within the V5 Flow Packet are translated to individual Metric objects.