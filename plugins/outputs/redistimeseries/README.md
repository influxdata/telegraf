# RedisTimeSeries Producer Output Plugin

The RedisTimeSeries output plugin writes metrics to the RedisTimeSeries server.

```toml
[[outputs.redistimeseries]]
  ## The address of the RedisTimeSeries server.
  addr = "127.0.0.1:6379"

  ## password to login Redis
  # password = ""

```

### Required parameters:

* `addr`: The address of the RedisTimeSeries server

### Optional parameters:
* `password`: The password to connect RedisTimeSeries server.
