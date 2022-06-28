# RedisTimeSeries Producer Output Plugin

The RedisTimeSeries output plugin writes metrics to the RedisTimeSeries server.

## Configuration

```toml
[[outputs.redistimeseries]]
  ## The address of the RedisTimeSeries server.
  address = "127.0.0.1:6379"
  ## password to login Redis
  password = ""

  ## username (optional)
  # username = ""
  # redis database number (optional, must be an integer)
  # database = 0

  ## optional TLS configurations
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  # insecure_skip_verify = false
```
