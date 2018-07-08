# NATS Streaming Input Plugin

The [NATS Streaming](https://nats.io/documentation/streaming/nats-streaming-intro/) monitoring plugin gathers metrics from
the NATS streaming system [monitoring http server](https://github.com/nats-io/nats-streaming-server#monitoring).

### Configuration

```toml
[[inputs.nats_streaming]]
  ## The address of the monitoring endpoint of the NATS Streaming server
  server = "http://localhost:8222"

  ## Maximum time to receive response
  # response_timeout = "5s"
```

### Metrics:

- nats_streaming_server
  - tags
    - server
    - cluster_id
    - server_id
  - fields:
    - clients (integer, count)
    - subscriptions (integer, count)
    - channels (integer, count)
    - total_msgs (integer, count)
    - total_bytes (integer, bytes)
    - uptime (integer, nanoseconds)

- nats_streaming_channel
  - tags
    - server
    - cluster_id
    - server_id
    - channel_name
  - fields:
    - msgs (integer, count)
    - bytes (integer, bytes)
    - first_seq (integer, count)
    - last_seq (integer, count)

- nats_streaming_subscription
  - tags
    - server
    - cluster_id
    - server_id
    - channel_name
    - client_id
    - inbox
    - ack_inbox
    - durable_name
    - queue_name
  - fields:
    - is_durable (bool, flag)
    - is_offline (bool, flag)
    - max_inflight (integer, count)
    - ack_wait (integer, count)
    - pending_count (integer, count)
    - is_stalled (bool, flag)

### Example Output:

```
nats_streaming_server,cluster_id=test-cluster,server_id=KX1Y9BA1M7cPjLhZ7rldxm,server=http://172.20.0.110:8222 channels=0i,clients=0i,subscriptions=0i,total_bytes=0i,total_msgs=0i,uptime=8767332199527i 1530898217000000000
```
