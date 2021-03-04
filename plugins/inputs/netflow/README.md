# Netflow Input Service Plugin

The Netflow plugin gathers network metrics

### Configuration:

```toml
# Netflow listener
[[inputs.netflow]]
  ## Address and port to host Netflow listener on
  service_address = ":2055"
  ## Number of Netflow messages allowed to queue up. Once filled, the
  ## Netflow listener will start dropping packets.
  allowed_pending_messages = 10000

  resolve_application_name_by_id = true
  resolve_ifname_by_ifindex = true
```

### Measurements & Fields:

#### Version 5

In Netflow version 5, 18 static fields as below.

- src_addr
- dst_addr
- nexthop
- input
- output
- packets
- bytes
- first
- last
- src_port
- dst_port
- tcp_flags
- protocol
- tos
- src_as
- dst_as
- src_mask
- dst_mask

#### Version 9

In Netflow version 9 or IPFIX, fields are defined dynamically by templates.

RFC3954-defined (RFC3954 Section 8) and Cisco-defined field types can be a field.

https://www.ietf.org/rfc/rfc3954.txt

**show flow exporter export-ids netflow-v9** command on Cisco IOS displays Cisco-defined field types.

#### IPFIX (Version 10)

IANA-registered IPFIX Information Elements and Cisco-defined field types can be a field.

http://www.iana.org/assignments/ipfix/ipfix.xhtml

**show flow exporter export-ids ipfix** command on Cisco IOS displays Cisco-defined field types. 

### Timestamp:

Currently, the localtime on the server where Telegraf runs is used for the timestamp.

### Resolve Application Name By ID

In case of Netflow version 9 or IPFIX, **application_name** tag was resolved by **application_id** field with following config.

```toml
# Netflow listener
[[inputs.netflow]]
  resolve_application_name_by_id = true
```

### Resolve Ifname By Ifindex

In case of Netflow version 9 or IPFIX, **interface_input_name** or **interface_output_name** was resolved by **interface_input_snmp** or **interface_output_snmp** with following config.

```toml
# Netflow listener
[[inputs.netflow]]
  resolve_ifname_by_ifindex = true
```