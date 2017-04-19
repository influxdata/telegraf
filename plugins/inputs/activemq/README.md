# Telegraf Input Plugin: ActiveMQ

This plugin gather queues, topics & subscribers metrics using ActiveMQ Console API.

### Configuration:

```toml
# Description
[[inputs.activemq]]
  ## Required ActiveMQ Endpoint
  server = "192.168.50.10"
  ## Required ActiveMQ port
  port = 8161
  ## Required username used for request HTTP Basic Authentication
  username = "admin"
  ## Required password used for HTTP Basic Authentication
  password = "admin"
  ## Required ActiveMQ webadmin root path
  webadmin = "admin"
```

### Measurements & Fields:

Every effort was made to preserve the names based on the XML response from the ActiveMQ Console API.

- queues_metrics:
    - size
    - consumer_count
    - enqueue_count
    - dequeue_count
  - topics_metrics:
    - size
    - consumer_count
    - enqueue_count
    - dequeue_count
  - subscribers_metrics:
    - pending_queue_size
    - dispatched_queue_size
    - dispatched_counter
    - enqueue_counter
    - dequeue_counter

### Tags:

- queues_metrics:
    - name
- topics_metrics:
    - name
- subscribers_metrics:
    - client_id
    - subscription_name
    - connection_id
    - destination_name
    - selector
    - active

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter activemq -test
queues_metrics,name=sandra,host=88284b2fe51b consumer_count=0i,enqueue_count=0i,dequeue_count=0i,size=0i 1492610703000000000
queues_metrics,name=Test,host=88284b2fe51b dequeue_count=0i,size=0i,consumer_count=0i,enqueue_count=0i 1492610703000000000
topics_metrics,name=ActiveMQ.Advisory.MasterBroker\ ,host=88284b2fe51b size=0i,consumer_count=0i,enqueue_count=1i,dequeue_count=0i 1492610703000000000
topics_metrics,host=88284b2fe51b,name=AAA\  size=0i,consumer_count=1i,enqueue_count=0i,dequeue_count=0i 1492610703000000000
topics_metrics,name=ActiveMQ.Advisory.Topic\ ,host=88284b2fe51b enqueue_count=1i,dequeue_count=0i,size=0i,consumer_count=0i 1492610703000000000
topics_metrics,name=ActiveMQ.Advisory.Queue\ ,host=88284b2fe51b size=0i,consumer_count=0i,enqueue_count=2i,dequeue_count=0i 1492610703000000000
topics_metrics,name=AAAA\ ,host=88284b2fe51b consumer_count=0i,enqueue_count=0i,dequeue_count=0i,size=0i 1492610703000000000
subscribers_metrics,connection_id=NOTSET,destination_name=AAA,selector=AA,active=no,host=88284b2fe51b,client_id=AAA,subscription_name=AAA pending_queue_size=0i,dispatched_queue_size=0i,dispatched_counter=0i,enqueue_counter=0i,dequeue_counter=0i 1492610703000000000
```
