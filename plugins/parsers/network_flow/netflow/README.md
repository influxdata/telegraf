# SFlow Parser

## Current Scope

Currently this Netflow Packet parser will  parse Netflow Version 10 (IPFIX) and as a by-product of the backwards compatability Netflow Version 9 as well.

# Schema
## Tags
| Name | Description |
|---|---|


## Fields
| Name | Type | Description |
|---|---|---|
|  bytes |  Integer | Derived from the product of frame_length and packets
|  packets | Integer |Decoded from sampling_rate field of flow_sample or flow_sample_expanded structures


## TO DO

When is cache of templates cleared for a particular source, could it be when we see a new
lower uptime?