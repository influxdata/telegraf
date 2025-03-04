# InfiniBand Input Plugin

This plugin gathers statistics for all InfiniBand devices and ports on the
system. These are the counters that can be found in
`/sys/class/infiniband/<dev>/port/<port>/counters/`
and RDMA counters can be found in
`/sys/class/infiniband/<dev>/ports/<port>/hw_counters/`

⭐ Telegraf v1.14.0
🏷️ network
💻 linux

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gets counters from all InfiniBand cards and ports installed
# This plugin ONLY supports Linux
[[inputs.infiniband]]
  # no configuration

  ## Collect RDMA counters
  # gather_rdma = false
```

## Metrics

Actual metrics depend on the InfiniBand devices, the plugin uses a simple
mapping from counter -> counter value.

[Information about the counters][counters] collected is provided by Nvidia.

[counters]: https://enterprise-support.nvidia.com/s/article/understanding-mlx5-linux-counters-and-status-parameters

The following fields are emitted by the plugin when selecting `counters`:

- infiniband
  - tags:
    - device
    - port
  - fields:

    ### Infiniband Counters

    - excessive_buffer_overrun_errors (integer)
    - link_downed (integer)
    - link_error_recovery (integer)
    - local_link_integrity_errors (integer)
    - multicast_rcv_packets (integer)
    - multicast_xmit_packets (integer)
    - port_rcv_constraint_errors (integer)
    - port_rcv_data (integer)
    - port_rcv_errors (integer)
    - port_rcv_packets (integer)
    - port_rcv_remote_physical_errors (integer)
    - port_rcv_switch_relay_errors (integer)
    - port_xmit_constraint_errors (integer)
    - port_xmit_data (integer)
    - port_xmit_discards (integer)
    - port_xmit_packets (integer)
    - port_xmit_wait (integer)
    - symbol_error (integer)
    - unicast_rcv_packets (integer)
    - unicast_xmit_packets (integer)
    - VL15_dropped (integer)

    ### Infiniband RDMA counters

    - duplicate_request (integer)
    - implied_nak_seq_err (integer)
    - lifespan (integer)
    - local_ack_timeout_err (integer)
    - np_cnp_sent (integer)
    - np_ecn_marked_roce_packets (integer)
    - out_of_buffer (integer)
    - out_of_sequence (integer)
    - packet_seq_err (integer)
    - req_cqe_error (integer)
    - req_cqe_flush_error (integer)
    - req_remote_access_errors (integer)
    - req_remote_invalid_request (integer)
    - resp_cqe_error (integer)
    - resp_cqe_flush_error (integer)
    - resp_local_length_error (integer)
    - resp_remote_access_errors (integer)
    - rnr_nak_retry_err (integer)
    - roce_adp_retrans (integer)
    - roce_adp_retrans_to (integer)
    - roce_slow_restart (integer)
    - roce_slow_restart_cnps (integer)
    - roce_slow_restart_trans (integer)
    - rp_cnp_handled (integer)
    - rp_cnp_ignored (integer)
    - rx_atomic_requests (integer)
    - rx_icrc_encapsulated (integer)
    - rx_read_requests (integer)
    - rx_write_requests (integer)

## Example Output

```text
infiniband,device=mlx5_bond_0,host=hop-r640-12,port=1 port_xmit_data=85378896588i,VL15_dropped=0i,port_rcv_packets=34914071i,port_rcv_data=34600185253i,port_xmit_discards=0i,link_downed=0i,local_link_integrity_errors=0i,symbol_error=0i,link_error_recovery=0i,multicast_rcv_packets=0i,multicast_xmit_packets=0i,unicast_xmit_packets=82002535i,excessive_buffer_overrun_errors=0i,port_rcv_switch_relay_errors=0i,unicast_rcv_packets=34914071i,port_xmit_constraint_errors=0i,port_rcv_errors=0i,port_xmit_wait=0i,port_rcv_remote_physical_errors=0i,port_rcv_constraint_errors=0i,port_xmit_packets=82002535i 1737652060000000000
infiniband,device=mlx5_bond_0,host=hop-r640-12,port=1 local_ack_timeout_err=0i,lifespan=10i,out_of_buffer=0i,resp_remote_access_errors=0i,resp_local_length_error=0i,np_cnp_sent=0i,roce_slow_restart=0i,rx_read_requests=6000i,duplicate_request=0i,resp_cqe_error=0i,rx_write_requests=19000i,roce_slow_restart_cnps=0i,rx_icrc_encapsulated=0i,rnr_nak_retry_err=0i,roce_adp_retrans=0i,out_of_sequence=0i,req_remote_access_errors=0i,roce_slow_restart_trans=0i,req_remote_invalid_request=0i,req_cqe_error=0i,resp_cqe_flush_error=0i,packet_seq_err=0i,roce_adp_retrans_to=0i,np_ecn_marked_roce_packets=0i,rp_cnp_handled=0i,implied_nak_seq_err=0i,rp_cnp_ignored=0i,req_cqe_flush_error=0i,rx_atomic_requests=0i 1737652060000000000
```
