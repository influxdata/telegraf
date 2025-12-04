# NSQ Input Plugin

This plugin gathers metrics from [NSQ][nsq] realtime distributed messaging
platform instances using the [NSQD API][api].

⭐ Telegraf v1.16.0
🏷️ server
💻 all

[nsq]: https://nsq.io/
[api]: https://nsq.io/components/nsqd.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read NSQ topic and channel statistics.
[[inputs.nsq]]
  ## An array of NSQD HTTP API endpoints
  endpoints  = ["http://localhost:4151"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Metrics

- `nsq_server`:
  - tags:
    - `server_host`
    - `server_version`
  - fields:
    - `server_count`
    - `topic_count`
- `nsq_topic`:
  - tags:
    - `server_host`
    - `server_version`
    - `topic`
  - fields:
    - `backend_depth`
    - `channel_count`
    - `depth`
    - `message_count`
- `nsq_channel`:
  - tags:
    - `server_host`
    - `server_version`
    - `topic`
    - `channel`
  - fields:
    - `backend_depth`
    - `client_count`
    - `depth`
    - `deferred_count`
    - `inflight_count`
    - `message_count`
    - `requeue_count`
    - `timeout_count`
- `nsq_client`:
  - tags:
    - `channel`
    - `client_address`
    - `client_hostname`
    - `client_id`
    - `client_name`
    - `client_user_agent`
    - `client_deflate`
    - `client_snappy`
    - `client_tls`
    - `client_version`
    - `server_host`
    - `server_version`
    - `topic`
  - fields:
    - `finish_count`
    - `inflight_count`
    - `message_count`
    - `ready_count`
    - `requeue_count`

## Example Output

```text
nsq_server,server_host=127.0.0.1:35871,server_version=0.3.6 server_count=1i,topic_count=2i 1742836824386224245
nsq_topic,server_host=127.0.0.1:35871,server_version=0.3.6,topic=t1 backend_depth=13i,channel_count=1i,depth=12i,message_count=14i 1742836824386235365
nsq_channel,channel=c1,server_host=127.0.0.1:35871,server_version=0.3.6,topic=t1 backend_depth=1i,client_count=1i,deferred_count=3i,depth=0i,inflight_count=2i,message_count=4i,requeue_count=5i,timeout_count=6i 1742836824386241985
nsq_client,channel=c1,client_address=172.17.0.11:35560,client_deflate=false,client_hostname=373a715cd990,client_id=373a715cd990,client_name=373a715cd990,client_snappy=false,client_tls=false,client_user_agent=nsq_to_nsq/0.3.6\ go-nsq/1.0.5,client_version=V2,server_host=127.0.0.1:35871,server_version=0.3.6,topic=t1 finish_count=9i,inflight_count=7i,message_count=8i,ready_count=200i,requeue_count=10i 1742836824386252905
nsq_topic,server_host=127.0.0.1:35871,server_version=0.3.6,topic=t2 backend_depth=29i,channel_count=1i,depth=28i,message_count=30i 1742836824386263806
nsq_channel,channel=c2,server_host=127.0.0.1:35871,server_version=0.3.6,topic=t2 backend_depth=16i,client_count=1i,deferred_count=18i,depth=15i,inflight_count=17i,message_count=19i,requeue_count=20i,timeout_count=21i 1742836824386270026
nsq_client,channel=c2,client_address=172.17.0.8:48145,client_deflate=true,client_hostname=377569bd462b,client_id=377569bd462b,client_name=377569bd462b,client_snappy=true,client_tls=true,client_user_agent=go-nsq/1.0.5,client_version=V2,server_host=127.0.0.1:35871,server_version=0.3.6,topic=t2 finish_count=25i,inflight_count=23i,message_count=24i,ready_count=22i,requeue_count=26i 1742836824386277926
```
