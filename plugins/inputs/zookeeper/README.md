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
- zookeeper_avg_latency
- zookeeper_max_latency
- zookeeper_min_latency
- zookeeper_packets_received
- zookeeper_packets_sent
- zookeeper_outstanding_requests
- zookeeper_znode_count
- zookeeper_watch_count
- zookeeper_ephemerals_count
- zookeeper_approximate_data_size
- zookeeper_followers #only exposed by the Leader
- zookeeper_synced_followers #only exposed by the Leader
- zookeeper_pending_syncs #only exposed by the Leader
- zookeeper_open_file_descriptor_count
- zookeeper_max_file_descriptor_count

#### Zookeeper string measurements:

Meta:
- units: string
- tags: `server=<hostname> port=<port>`

Measurement names:
- zookeeper_version
- zookeeper_server_state