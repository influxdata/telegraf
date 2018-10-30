# IPVS Input Plugin

The IPVS input plugin uses the linux kernel netlink socket interface to gather
metrics about ipvs virtual and real servers.

**Supported Platforms:** Linux

## Configuration

```toml
[[inputs.ipvs]]
  # no configuration
```

## Permissions

Assuming you installed the telegraf package via one of the published packages,
the process will be running as the `telegraf` user. However, in order for this
plugin to communicate over netlink sockets it needs the telegraf process to be
running as `root` (or some user with `CAP_NET_ADMIN` and `CAP_NET_RAW`). Be sure
to ensure these permissions before running telegraf with this plugin included.

## Example Output

### Virtual servers

For virtual servers, this plugin reports the following:

- Connections
- PacketsIn
- PacketsOut
- BytesIn
- BytesOut
- CPS
- PPSIn
- PPSOut
- BPSIn
- BPSOut

Each virtual server will contain tags identifying how it was configured, using one of:

- `address` + `port` + `protocol`
- `fwmark`

Example:
```
ipvs_virtual_server,address=172.18.64.234,address_family=inet,netmask=32,port=9000,protocol=tcp,sched=mh_418 bytes_out=0i,pps_in=0i,pps_out=0i,cps=0i,pkts_in=0i,pkts_out=0i,connections=0i,bytes_in=0i 1540407540000000000
ipvs_virtual_server,address_family=inet,fwmark=47,netmask=32,sched=mh_418 connections=0i,pkts_in=0i,bytes_out=0i,pps_in=0i,pps_out=0i,pkts_out=0i,bytes_in=0i,cps=0i 1540407540000000000
```

### Real servers

Metrics reported for each Real Server:

- ActiveConnections
- InactiveConnections
- Connections
- PacketsIn
- PacketsOut
- BytesIn
- BytesOut
- CPS
- PPSIn
- PPSOut
- BPSIn
- BPSOut

Each real server can be identified as belonging to a virtual server using one of:

- `virtual_address` + `virtual_port` + `virtual_protocol`
- `virtual_fwmark`

Example:
```
ipvs_real_server,active_conns=0,address=172.18.64.204,address_family=inet,inactive_conns=0,port=9000,virtual_address=172.18.64.234,virtual_port=9000,virtual_protocol=tcp bytes_out=0i,pps_in=0i,pps_out=0i,cps=0i,connections=0i,pkts_in=0i,pkts_out=0i,bytes_in=0i 1540577170000000000
ipvs_real_server,active_conns=0,address=172.18.64.203,address_family=inet,inactive_conns=0,port=9000,virtual_address=172.18.64.234,virtual_port=9000,virtual_protocol=tcp pps_out=0i,cps=0i,connections=0i,pkts_in=0i,pkts_out=0i,bytes_in=0i,bytes_out=0i,pps_in=0i 1540577170000000000
ipvs_real_server,active_conns=0,address=172.18.64.202,address_family=inet,inactive_conns=0,port=9000,virtual_address=172.18.64.234,virtual_port=9000,virtual_protocol=tcp pkts_in=0i,pkts_out=0i,bytes_in=0i,bytes_out=0i,pps_in=0i,pps_out=0i,cps=0i,connections=0i 1540577170000000000
ipvs_real_server,active_conns=0,address=172.18.64.201,address_family=inet,inactive_conns=0,port=9000,virtual_address=172.18.64.234,virtual_port=9000,virtual_protocol=tcp pps_in=0i,pps_out=0i,cps=0i,connections=0i,pkts_in=0i,pkts_out=0i,bytes_in=0i,bytes_out=0i 1540577170000000000
```
