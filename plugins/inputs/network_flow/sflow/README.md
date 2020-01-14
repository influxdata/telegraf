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
```

# Schema

The parsing of SFlow packets is handled by the SFlow Parser and the schema is described [here](../../parsers/network_flow/sflow/README.md).

At a high level, individual Flow Samples within the V5 Flow Packet are translated to individual Metric objects.


