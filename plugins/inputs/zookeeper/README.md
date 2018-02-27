## Telegraf Plugin: Zookeeper

#### Description

The zookeeper plugin collects variables outputted from the 'mntr' command
[Zookeeper Admin](https://zookeeper.apache.org/doc/trunk/zookeeperAdmin.html).

```
echo mntr | nc localhost 2181

              zk_version  3.4.0
              zk_avg_latency  0
              zk_max_latency  0
              zk_min_latency  0
              zk_packets_received 70
              zk_packets_sent 69
              zk_outstanding_requests 0
              zk_server_state leader
              zk_znode_count   4
              zk_watch_count  0
              zk_ephemerals_count 0
              zk_approximate_data_size    27
              zk_followers    4                   - only exposed by the Leader
              zk_synced_followers 4               - only exposed by the Leader
              zk_pending_syncs    0               - only exposed by the Leader
              zk_open_file_descriptor_count 23    - only available on Unix platforms
              zk_max_file_descriptor_count 1024   - only available on Unix platforms
```

## Configuration

```
# Reads 'mntr' stats from one or many zookeeper servers
[[inputs.zookeeper]]
  ## An array of address to gather stats about. Specify an ip or hostname
  ## with port. ie localhost:2181, 10.0.0.1:2181, etc.

  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 2181 is used
  servers = [":2181"]
```

## InfluxDB Measurement:

```
M zookeeper
  T host
  T port
  T state
  
  F approximate_data_size        integer
  F avg_latency                  integer
  F ephemerals_count             integer
  F max_file_descriptor_count    integer
  F max_latency                  integer
  F min_latency                  integer
  F num_alive_connections        integer
  F open_file_descriptor_count   integer
  F outstanding_requests         integer
  F packets_received             integer
  F packets_sent                 integer
  F version                      string
  F watch_count                  integer
  F znode_count                  integer
```