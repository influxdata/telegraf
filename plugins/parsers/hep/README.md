# HEP

The HEP data format parses a HEP packet into metric fields.

**NOTE:** All HEP packets are stores as Tags unless provided specifically
provided with `hep_header` array and body is parsed with JSON parser.
All the JSON parser features were imported in Hep parser. Please check Json Parser for more details. 

Any field/header can be ignored using already telegraf filters. 

### Configuration

```toml
[[inputs.socket_listener]]
  service_address = "udp://:8094"
  data_format = "hep"
  hep_header = ["protocol"]

```
