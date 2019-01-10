# RedisTS Producer Output Plugin

The RedisTS output plugin writes metrics to the RedisTS server.

```toml
[[outputs.redists]]
  ## The address of the RedisTS server.
  addr = "127.0.0.1:6379"

  ## password to login Redis
  # password = ""

```

### Required parameters:

* `addr`: The address of the RedisTS server

### Optional parameters:
* `password`: The password to connect MQTT server.
