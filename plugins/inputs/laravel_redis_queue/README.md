# Telegraf Plugin: Laravel Redis Queue

### Configuration:

```
# Read Redis's basic status information
[[inputs.laravel_redis_queue]]
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://:password@192.168.99.100
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]

  ## specify queues:
	##  [queue_name]
	##  e.g.
	##    queue1
	queues = ["queue1", "queue2"]
```

### Measurements & Fields:

- Measurement
    - delayed_count_queue_name
    - pushed_count_queue_name
    - reserved_count_queue_name

### Tags:

- All measurements have the following tags:
    - host
    - port
    - server

### Example Output:

Using this configuration:
```
[[inputs.laravel_redis_queue]]
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://:password@192.168.99.100
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]

  ## specify queues:
	##  [queue_name]
	##  e.g.
	##    queue1
	queues = ["default"]
```

When run with:
```
./telegraf -config telegraf.conf -input-filter laravel_redis_queue -test
```

It produces:
```
* Plugin: inputs.laravel_redis_queue, Collection 1
> laravel_redis_queue,host=luoxiaojun1992-OptiPlex-7020,port=6379,server=192.168.169.13 pushed_count_default=0,delayed_count_default=1,reserved_count_default=0 1480145343000000000
```
