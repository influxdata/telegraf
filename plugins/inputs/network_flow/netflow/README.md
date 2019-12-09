# SFlow Input Plugin

The Netflow Input Plugin provides support for acting as an Netflow V9/V10 collector in accordance with the specification from [IETF](https://tools.ietf.org/html/rfc7011).


# Configuration
The following configuration options are availabe:

| Name | Description 
|---|---|
| service_address| URL to listen on expressed as UDP (IPv4 or 6) OP address and port number 
| | Example: ```service_address = "udp://:2055"```
| read_buffer_size | Maximum socket buffer size (in bytes when no unit specified). Once the buffer fills up, metrics will start dropping. Defaults to the OS default.
||Example = ```read_buffer_size"64KiB"``` |
| dns_multi_name_processor | An optional regexp and template to use to transform a DNS resolve name. Particularily useful when DNS resolves an IP address to more than one name, and they alternative in order when queried. Using this processor command it is possible to tranform the name into something common irrespect of which entry is first - if the names conform to a regular naming schema. Note TOML [escape sequences](https://github.com/toml-lang/toml) may be required.
||Example: ````s/(.*)(?:-net[0-9])/$1```` will strip ```-net<n>``` from the host name thereby converting, as an example, ```hostx-net1``` and ```hostx-net2``` both to ```hostx```
|dns_fqdn_resolve|Determines whether IP addresses should be resolved to Host names.
||Example: ```dns_fqdn_resolve = true```
|dns_fqdn_cache_ttl|The time to live for entries in the DNS name cache expressed in seconds. Default is 0 which is infinite
||Example: ```dns_fwdn_cache_ttl = 3600```

## Configuration:

This is a sample configuration for the plugin.

```toml
[[inputs.netflow]]
	## URL to listen on
	# service_address = "udp://:2055"
	# service_address = "udp4://:2055"
	# service_address = "udp6://:2055"
    
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
```

## DNS Name and SNMP Interface name resolution and caching

Raw Netflow packets, and their sample data, communicate IP addresses which are not very useful to humans.

The Netflow plugin can be configured to attempt to resolve IP addresses to host names via DNS.

The resolved names, or in the case of a resolution error the ip/id will be used as 'the' name, are configurably cached for a period of time to avoid continual lookups.

| Source IP Tag | Resolved Host Tag 
|---|---|
|agentAddress|agentHost|
|sourceIPv4Address|sourceIPv4Host|
|destinationIPv4Address|sourceIPv4Host|
|sourceIPv6Address|sourceIPv6Host|
|destinationIPv6Address|destinationIPv6Host|
|exporterIPv4Address|exporterIPv4Host|
|exporterIPv6Address|exporterIPv6Host|


### Multipe DNS Name resolution & processing

In some cases DNS servers may maintain multiple entries for the same IP address in support of load balancing. In this setup the same IP address may be resolved to multiple DNS names, via a single DNS query, and it is likely the order of those DNS names will change over time.

In order to provide some stability to the names recorded against flow records, it is possible to provide a regular expression and template transformation that should be capable of converting multiple names to a single common name where a mathodical naming scheme has been used.

Example: ````s/(.*)(?:-net[0-9])/$1```` will strip ```-net<n>``` from the host name thereby converting, as an example, ```hostx-net1``` and ```hostx-net2``` both to ```hostx```

# Schema

The parsing of Netflow packets is handled by the Netflow Parser and the schema is described [here](../../parsers/network_flow/netflow/README.md).

At a high level, individual Flow Samples within the V10 Flow Packet are translated to individual Metric objects.


