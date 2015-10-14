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

## Measurements:
#### Zookeeper measurements:

Meta:
- units: int64
- tags: `server=<hostname> port=<port>`

Measurement names:
- avg_latency
- max_latency
- min_latency
- packets_received
- packets_sent
- outstanding_requests
- znode_count
- watch_count
- ephemerals_count
- approximate_data_size
- followers #only exposed by the Leader
- synced_followers #only exposed by the Leader
- pending_syncs #only exposed by the Leader
- open_file_descriptor_count
- max_file_descriptor_count

#### Zookeeper string measurements:

Meta:
- units: string
- tags: `server=<hostname> port=<port>`

Measurement names:
- zk_version
- server_state