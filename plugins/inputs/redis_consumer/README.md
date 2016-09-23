# Redis Consumer Input Plugin

The [Redis](http://http://redis.io//) consumer plugin subscribes to one or more
 Redis channels and adds messages to InfluxDB. Multiple Redis servers may be specified
 at a time. The Redis consumer may be configured to use both standard channel names or
 patterned channel names.

## Configuration

```toml
# Read metrics from Redis channel(s)
[[inputs.redis_consumer]]
  servers = ["tcp://localhost:6379"]

  ##  List of channels to listen to. Selecting channels using Redis'
  ##  pattern-matching is allowed, e.g.:
  ##	channels = ["telegraf:*", "app_[1-3]"]
  ##
  ##  See http://redis.io/topics/pubsub#pattern-matching-subscriptions for
  ##  more info.
  channels = ["telegraf"]

  ## Data format to consume. This can be "json", "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## Testing

Running integration tests requires running Redis. See Makefile
for redis container command.
