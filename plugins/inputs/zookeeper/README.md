# Zookeeper Input Plugin

The zookeeper plugin collects variables outputted from the 'mntr' command
[Zookeeper Admin](https://zookeeper.apache.org/doc/current/zookeeperAdmin.html).

## Configuration

```toml @sample.conf
# Reads metrics from one or many zookeeper servers
[[inputs.zookeeper]]
  ## An array of address to gather stats about. Specify an ip or hostname
  ## with port. ie localhost:2181, 10.0.0.1:2181, etc.
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 2181 is used for java and 7000 for prometheus
  ## metric providers.
  servers = [":2181"]

  ## Timeout for metric collections from all servers.  Minimum timeout is "1s".
  # timeout = "5s"

  ## Metrics Provider
  ## Choose from: "java" or "prometheus". By default, mntr is used to collect
  ## metrics produced in the Java Properties format. There is the option to
  ## also produce Prometheus style metrics from Zookeeper. This requires
  ## additional configuraiton. Using this provider requires the use of a
  ## different port, 7000 by default, and will produce metrics in a different
  ## format.
  # metrics_provider = "java"

  ## Optional TLS Config
  # enable_tls = true
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## If false, skip chain & host verification
  # insecure_skip_verify = true
```

## Metrics

Exact field names are based on Zookeeper response and may vary between
configuration, platform, and version.

- zookeeper
  - tags:
    - server
    - port
    - state
  - fields:
    - approximate_data_size (integer)
    - avg_latency (integer)
    - ephemerals_count (integer)
    - max_file_descriptor_count (integer)
    - max_latency (integer)
    - min_latency (integer)
    - num_alive_connections (integer)
    - open_file_descriptor_count (integer)
    - outstanding_requests (integer)
    - packets_received (integer)
    - packets_sent (integer)
    - version (string)
    - watch_count (integer)
    - znode_count (integer)
    - followers (integer, leader only)
    - synced_followers (integer, leader only)
    - pending_syncs (integer, leader only)

## Debugging

If you have any issues please check the direct Zookeeper output using netcat:

```sh
$ echo mntr | nc localhost 2181
zk_version      3.4.9-3--1, built on Thu, 01 Jun 2017 16:26:44 -0700
zk_avg_latency  0
zk_max_latency  0
zk_min_latency  0
zk_packets_received     8
zk_packets_sent 7
zk_num_alive_connections        1
zk_outstanding_requests 0
zk_server_state standalone
zk_znode_count  129
zk_watch_count  0
zk_ephemerals_count     0
zk_approximate_data_size        10044
zk_open_file_descriptor_count   44
zk_max_file_descriptor_count    4096
```

### Prometheus Metric Provider

The Prometheus Metric provider does not use `mntr` and instead pulls down
metrics from the exposed Prometheus `/metrics` endpoint on port 7000, by
default. Users should be able to `curl` or open in a browser this endpoint and
verify that metrics are generated if the correct Zookeeper configuration
settings are enabled.

## Example Output

With the default, "java" metric provider:

```shell
zookeeper,server=localhost,port=2181,state=standalone ephemerals_count=0i,approximate_data_size=10044i,open_file_descriptor_count=44i,max_latency=0i,packets_received=7i,outstanding_requests=0i,znode_count=129i,max_file_descriptor_count=4096i,version="3.4.9-3--1",avg_latency=0i,packets_sent=6i,num_alive_connections=1i,watch_count=0i,min_latency=0i 1522351112000000000
```

With the "prometheus" metric provider:

```shell
prometheus,source=localhost:7000 read_commit_proc_issued_count=0,read_commit_proc_issued_sum=0 1659986918000000000
prometheus,source=localhost:7000 proposal_latency_count=0,proposal_latency_sum=0 1659986918000000000
prometheus,source=localhost:7000 max_file_descriptor_count=1048576 1659986918000000000
prometheus,source=localhost:7000 reads_after_write_in_session_queue_count=0,reads_after_write_in_session_queue_sum=0 1659986918000000000
prometheus,source=localhost:7000 commit_process_time_count=0,commit_process_time_sum=0 1659986918000000000
prometheus,source=localhost:7000 om_commit_process_time_ms_count=0,om_commit_process_time_ms_sum=0 1659986918000000000
prometheus,source=localhost:7000 packets_sent=0 1659986918000000000
prometheus,source=localhost:7000 outstanding_tls_handshake=0 1659986918000000000
prometheus,source=localhost:7000 packets_received=0 1659986918000000000
prometheus,source=localhost:7000 pending_session_queue_size_count=0,pending_session_queue_size_sum=0 1659986918000000000
prometheus,pool=mapped,source=localhost:7000 jvm_buffer_pool_used_buffers=0 1659986918000000000
prometheus,pool=direct,source=localhost:7000 jvm_buffer_pool_used_buffers=23 1659986918000000000
prometheus,source=localhost:7000 quit_leading_due_to_disloyal_voter=0 1659986918000000000
prometheus,source=localhost:7000 sync_processor_queue_and_flush_time_ms_count=0,sync_processor_queue_and_flush_time_ms_sum=0 1659986918000000000
...
```
