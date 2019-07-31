# RabbitMQ Input Plugin

Reads metrics from RabbitMQ servers via the [Management Plugin](https://www.rabbitmq.com/management.html).

For additional details reference the [RabbitMQ Management HTTP Stats](https://cdn.rawgit.com/rabbitmq/rabbitmq-management/master/priv/www/doc/stats.html).

### Configuration:

```toml
[[inputs.rabbitmq]]
  ## Management Plugin url. (default: http://localhost:15672)
  # url = "http://localhost:15672"
  ## Tag added to rabbitmq_overview series; deprecated: use tags
  # name = "rmq-server-1"
  ## Credentials
  # username = "guest"
  # password = "guest"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional request timeouts
  ##
  ## ResponseHeaderTimeout, if non-zero, specifies the amount of time to wait
  ## for a server's response headers after fully writing the request.
  # header_timeout = "3s"
  ##
  ## client_timeout specifies a time limit for requests made by this client.
  ## Includes connection time, any redirects, and reading the response body.
  # client_timeout = "4s"

  ## A list of nodes to gather as the rabbitmq_node measurement. If not
  ## specified, metrics for all nodes are gathered.
  # nodes = ["rabbit@node1", "rabbit@node2"]

  ## A list of queues to gather as the rabbitmq_queue measurement. If not
  ## specified, metrics for all queues are gathered.
  # queues = ["telegraf"]

  ## A list of exchanges to gather as the rabbitmq_exchange measurement. If not
  ## specified, metrics for all exchanges are gathered.
  # exchanges = ["telegraf"]

  ## Queues to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all queues
  # queue_name_include = []
  # queue_name_exclude = []
```

### Measurements & Fields:

- rabbitmq_overview
  - channels (int, channels)
  - connections (int, connections)
  - consumers (int, consumers)
  - exchanges (int, exchanges)
  - messages (int, messages)
  - messages_acked (int, messages)
  - messages_delivered (int, messages)
  - messages_delivered_get (int, messages)
  - messages_published (int, messages)
  - messages_ready (int, messages)
  - messages_unacked (int, messages)
  - queues (int, queues)
  - clustering_listeners (int, cluster nodes)
  - amqp_listeners (int, amqp nodes up)
  - return_unroutable (int, number of unroutable messages)
  - return_unroutable_rate (float, number of unroutable messages per second)

- rabbitmq_node
  - disk_free (int, bytes)
  - disk_free_limit (int, bytes)
  - disk_free_alarm (int, disk alarm)
  - fd_total (int, file descriptors)
  - fd_used (int, file descriptors)
  - mem_limit (int, bytes)
  - mem_used (int, bytes)
  - mem_alarm (int, memory a)
  - proc_total (int, erlang processes)
  - proc_used (int, erlang processes)
  - run_queue (int, erlang processes)
  - sockets_total (int, sockets)
  - sockets_used (int, sockets)
  - running (int, node up)
  - uptime (int, milliseconds)
  - health_check_status (int, 1 or 0)
  - mnesia_disk_tx_count (int, number of disk transaction)
  - mnesia_ram_tx_count (int, number of ram transaction)
  - mnesia_disk_tx_count_rate (float, number of disk transaction per second)
  - mnesia_ram_tx_count_rate (float, number of ram transaction per second)
  - gc_num (int, number of garbage collection)
  - gc_bytes_reclaimed (int, bytes)
  - gc_num_rate (float, number of garbage collection per second)
  - gc_bytes_reclaimed_rate (float, bytes per second)
  - io_read_avg_time (float, number of read operations)
  - io_read_avg_time_rate (int, number of read operations per second)
  - io_read_bytes (int, bytes)
  - io_read_bytes_rate (float, bytes per second)
  - io_write_avg_time (int, milliseconds)
  - io_write_avg_time_rate (float, milliseconds per second)
  - io_write_bytes (int, bytes)
  - io_write_bytes_rate (float, bytes per second)

- rabbitmq_queue
  - consumer_utilisation (float, percent)
  - consumers (int, int)
  - idle_since (string, time - e.g., "2006-01-02 15:04:05")
  - memory (int, bytes)
  - message_bytes (int, bytes)
  - message_bytes_persist (int, bytes)
  - message_bytes_ram (int, bytes)
  - message_bytes_ready (int, bytes)
  - message_bytes_unacked (int, bytes)
  - messages (int, count)
  - messages_ack (int, count)
  - messages_ack_rate (float, messages per second)
  - messages_deliver (int, count)
  - messages_deliver_rate (float, messages per second)
  - messages_deliver_get (int, count)
  - messages_deliver_get_rate (float, messages per second)
  - messages_publish (int, count)
  - messages_publish_rate (float, messages per second)
  - messages_ready (int, count)
  - messages_redeliver (int, count)
  - messages_redeliver_rate (float, messages per second)
  - messages_unack (integer, count)

- rabbitmq_exchange
  - messages_publish_in (int, count)
  - messages_publish_in_rate (int, messages per second)
  - messages_publish_out (int, count)
  - messages_publish_out_rate (int, messages per second)

### Tags:

- All measurements have the following tags:
  - url

- rabbitmq_overview
  - name

- rabbitmq_node
  - node
  - url

- rabbitmq_queue
  - url
  - queue
  - vhost
  - node
  - durable
  - auto_delete

- rabbitmq_exchange
  - url
  - exchange
  - type
  - vhost
  - internal
  - durable
  - auto_delete

### Sample Queries:

Message rates for the entire node can be calculated from total message counts. For instance, to get the rate of messages published per minute, use this query:

```
SELECT NON_NEGATIVE_DERIVATIVE(LAST("messages_published"), 1m) AS messages_published_rate
FROM rabbitmq_overview WHERE time > now() - 10m GROUP BY time(1m)
```

### Example Output:

```
rabbitmq_queue,url=http://amqp.example.org:15672,queue=telegraf,vhost=influxdb,node=rabbit@amqp.example.org,durable=true,auto_delete=false,host=amqp.example.org messages_deliver_get=0i,messages_publish=329i,messages_publish_rate=0.2,messages_redeliver_rate=0,message_bytes_ready=0i,message_bytes_unacked=0i,messages_deliver=329i,messages_unack=0i,consumers=1i,idle_since="",messages=0i,messages_deliver_rate=0.2,messages_deliver_get_rate=0.2,messages_redeliver=0i,memory=43032i,message_bytes_ram=0i,messages_ack=329i,messages_ready=0i,messages_ack_rate=0.2,consumer_utilisation=1,message_bytes=0i,message_bytes_persist=0i 1493684035000000000
rabbitmq_overview,url=http://amqp.example.org:15672,host=amqp.example.org channels=2i,consumers=1i,exchanges=17i,messages_acked=329i,messages=0i,messages_ready=0i,messages_unacked=0i,connections=2i,queues=1i,messages_delivered=329i,messages_published=329i,clustering_listeners=2i,amqp_listeners=1i 1493684035000000000
rabbitmq_node,url=http://amqp.example.org:15672,node=rabbit@amqp.example.org,host=amqp.example.org fd_total=1024i,fd_used=32i,mem_limit=8363329126i,sockets_total=829i,disk_free=8175935488i,disk_free_limit=50000000i,mem_used=58771080i,proc_total=1048576i,proc_used=267i,run_queue=0i,sockets_used=2i,running=1i 149368403500000000
rabbitmq_exchange,url=http://amqp.example.org:15672,exchange=telegraf,type=fanout,vhost=influxdb,internal=false,durable=true,auto_delete=false,host=amqp.example.org messages_publish_in=2i,messages_publish_out=1i 149368403500000000
```
