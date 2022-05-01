# Port Name Lookup Processor Plugin

Use the `port_name` processor to convert a tag or field containing a well-known port number to the registered service name.

Tag or field can contain a number ("80") or number and protocol separated by slash ("443/tcp"). If protocol is not provided it defaults to tcp but can be changed with the default_protocol setting. An additional tag or field can be specified for the protocol.

If the source was found in tag, the service name will be added as a tag. If the source was found in a field, the service name will also be a field.

Telegraf minimum version: Telegraf 1.15.0

## Configuration

```toml
# Given a tag/field of a TCP or UDP port number, add a tag/field of the service name looked up in the system services file
[[processors.port_name]]
  ## Name of tag holding the port number
  # tag = "port"
  ## Or name of the field holding the port number
  # field = "port"

  ## Name of output tag or field (depending on the source) where service name will be added
  # dest = "service"

  ## Default tcp or udp
  # default_protocol = "tcp"

  ## Tag containing the protocol (tcp or udp, case-insensitive)
  # protocol_tag = "proto"

  ## Field containing the protocol (tcp or udp, case-insensitive)
  # protocol_field = "proto"
```

## Example

```diff
- measurement,port=80 field=123 1560540094000000000
+ measurement,port=80,service=http field=123 1560540094000000000
```
