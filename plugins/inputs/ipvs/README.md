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

## Metrics

### Virtual Servers

Metrics report for each `ipvs_virtual_server`:

- `ipvs_virtual_server`
  - tags:
    - `sched` - the scheduler in use
    - `netmask` - the mask used for determining affinity
    - `address_family` - inet/inet6
    - ONE of `address` + `port` + `protocol` *OR* `fwmark`
  - fields:
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

Each virtual server will contain tags identifying how it was configured, using
one of `address` + `port` + `protocol` *OR* `fwmark`. This is how one would
normally configure a virtual server using `ipvsadm`.

### Real Servers

Metrics reported for each `ipvs_real_server`:

- `ipvs_real_server`
  - tags:
    - `address`
    - `port`
    - `address_family`
    - ONE of `virtual_address` + `virtual_port` + `virtual_protocol` OR `virtual_fwmark`
  - fields:
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

Each real server can be identified as belonging to a virtual server using one of
either `virtual_address + virtual_port + virtual_protocol` OR `virtual_fwmark`

## Example Output

### Virtual servers

Example (when a virtual server is configured using `fwmark` and backed by 2 real servers):
```
ipvs_virtual_server,address=172.18.64.234,address_family=inet,netmask=32,port=9000,protocol=tcp,sched=rr bytes_in=0i,bytes_out=0i,pps_in=0i,pps_out=0i,cps=0i,connections=0i,pkts_in=0i,pkts_out=0i 1541019340000000000
ipvs_real_server,address=172.18.64.220,address_family=inet,port=9000,virtual_address=172.18.64.234,virtual_port=9000,virtual_protocol=tcp active_connections=0i,inactive_connections=0i,pkts_in=0i,bytes_out=0i,pps_out=0i,connections=0i,pkts_out=0i,bytes_in=0i,pps_in=0i,cps=0i 1541019340000000000
ipvs_real_server,address=172.18.64.219,address_family=inet,port=9000,virtual_address=172.18.64.234,virtual_port=9000,virtual_protocol=tcp active_connections=0i,inactive_connections=0i,pps_in=0i,pps_out=0i,connections=0i,pkts_in=0i,pkts_out=0i,bytes_in=0i,bytes_out=0i,cps=0i 1541019340000000000

```

### Real servers

Example (when a real server is configured using `proto+addr+port` and backed by 2 real servers):
```
ipvs_virtual_server,address_family=inet,fwmark=47,netmask=32,sched=rr cps=0i,connections=0i,pkts_in=0i,pkts_out=0i,bytes_in=0i,bytes_out=0i,pps_in=0i,pps_out=0i 1541019340000000000
ipvs_real_server,address=172.18.64.220,address_family=inet,port=9000,virtual_fwmark=47 inactive_connections=0i,pkts_out=0i,bytes_out=0i,pps_in=0i,cps=0i,active_connections=0i,pkts_in=0i,bytes_in=0i,pps_out=0i,connections=0i 1541019340000000000
ipvs_real_server,address=172.18.64.219,address_family=inet,port=9000,virtual_fwmark=47 cps=0i,active_connections=0i,inactive_connections=0i,connections=0i,pkts_in=0i,bytes_out=0i,pkts_out=0i,bytes_in=0i,pps_in=0i,pps_out=0i 1541019340000000000
```
