# Port Name Lookup Processor Plugin

Use the `port_name` processor to convert a tag containing a well-known port number to the registered service name.

Tag can contain a number ("80") or number and protocol separated by slash ("443/tcp"). If protocol is not provided it defaults to tcp but can be changed with the default_protocol setting.

### Configuration

```toml
[[processors.port_name]]
  ## Name of tag holding the port number
  # tag = "port"

  ## Name of output tag where service name will be added
  # dest = "service"

  ## Default tcp or udp
  # default_protocol = "tcp"
```

### Example

```diff
- measurement,port=80 field=123 1560540094000000000
+ measurement,port=80,service=http field=123 1560540094000000000
```
