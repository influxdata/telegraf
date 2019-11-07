# InfiniBand Input Plugin

This plugin gathers statistics for all InfiniBand devices and ports on the system. These are the counters that can be found in /sys/class/infiniband/<dev>/port/<port>/counters/

### Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage infiniband`.

```toml
[[inputs.infiniband]]
```

There are no configuration options for this plugin.

### Metrics

You can find more information about the counters that are gathered here: 
https://community.mellanox.com/s/article/understanding-mlx5-linux-counters-and-status-parameters

There is a simple mapping from counter -> counter value. All counter values are 64 bit integers. A seperate measurement is made for each port.
Each measurement is tagged with the device and port that it relates to. These are strings.


### Example Output

```
infiniband,device=mlx5_0,port=1,VL15_dropped=0i,excessive_buffer_overrun_errors=0i,link_downed=0i,link_error_recovery=0i,local_link_integrity_errors=0i,multicast_rcv_packets=0i,multicast_xmit_packets=0i,port_rcv_constraint_errors=0i,port_rcv_data=237159415345822i,port_rcv_errors=0i,port_rcv_packets=801977655075i,port_rcv_remote_physical_errors=0i,port_rcv_switch_relay_errors=0i,port_xmit_constraint_errors=0i,port_xmit_data=238334949937759i,port_xmit_discards=0i,port_xmit_packets=803162651391i,port_xmit_wait=4294967295i,symbol_error=0i,unicast_rcv_packets=801977655075i,unicast_xmit_packets=803162651391i 1573125558000000000
```
