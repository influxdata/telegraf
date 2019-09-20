# SFlow Input Plugin

The SFlow Input Plugin provides support for acting as an SFlow V5 collector in accordance with the specification from [sflow.org](https://sflow.org/).

Currently only Flow Samples of Ethernet / IPv4 & IPv4 TCP & UDP headers are turned into metrics - counters and other header samples may come later.

# Configuration
The following configuration options are availabe:

| Name | Description 
|---|---|
| service_address| URL to listen on expressed as UDP (IPv4 or 6) OP address and port number 
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

The parsing of SFlow packets is handled by the SFlow Parser and the schema is described [here](../../parsers/sflow/README.md).

At a high level, individual Flow Samples within the V5 Flow Packet are translated to individual Metric objects.

## Tags (optionally as Fields)

The following items are naturally recorded as tags by the sflow plugin. However, using the ```as_fields``` configuration parameter it is possible to have any of these recorded as fields instead