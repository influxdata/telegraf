# IPVS Input Plugin (Linux)

The IPVS input plugin uses the linux kernel netlink socket interface to gather
metrics about ipvs virtual and real servers.

## Configuration

None

## Permissions

Assuming you installed the telegraf package via one of the published packages,
the process will be running as the `telegraf` user. However, in order for this
plugin to communicate over netlink sockets it needs the telegraf process to be
running as `root` (or some user with `CAP_NET_ADMIN` and `CAP_NET_RAW`). Be sure
to ensure these permissions before running telegraf with this plugin included.

## Sample Output

This is what you can expect the emitted metrics to look like

```
{"fields":{"bytes_in":0,"bytes_out":0,"connections":0,"cps":0,"pkts_in":0,"pkts_out":0,"pps_in":0,"pps_out":0},"name":"ipvs_virtual_server","tags":{"address_family":"inet","netmask":"32","sched":"rr","address":"172.18.64.234","port":"9000","protocol":"tcp"},"timestamp":1539810710}
{"fields":{"bytes_in":0,"bytes_out":0,"connections":0,"cps":0,"pkts_in":0,"pkts_out":0,"pps_in":0,"pps_out":0},"name":"ipvs_virtual_server","tags":{"address_family":"inet","netmask":"32","sched":"rr","fwmark":"47"},"timestamp":1539810710}
```
