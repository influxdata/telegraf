# InfluxDB Output Plugin

This plugin writes to [InfluxDB](https://www.influxdb.com) via HTTP or UDP.

### Configuration:

```toml
# Configuration for influxdb server to send metrics to
[[outputs.influxdb]]
  ## The HTTP or UDP URL for your InfluxDB instance.  Each item should be
  ## of the form:
  ##   scheme "://" host [ ":" port]
  ##
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  # urls = ["udp://localhost:8089"] # UDP endpoint example
  urls = ["http://localhost:8086"] # required
  ## The target database for metrics (telegraf will create it if not exists).
  database = "telegraf" # required

  ## Name of existing retention policy to write to.  Empty string writes to
  ## the default retention policy.
  retention_policy = ""
  ## Write consistency (clusters only), can be: "any", "one", "quorum", "all"
  write_consistency = "any"

  ## Write timeout (for the InfluxDB client), formatted as a string.
  ## If not provided, will default to 5s. 0s means no timeout (not recommended).
  timeout = "5s"
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"
  ## Set the user agent for HTTP POSTs (can be useful for log differentiation)
  # user_agent = "telegraf"
  ## Set UDP payload size, defaults to InfluxDB UDP Client default (512 bytes)
  # udp_payload = 512

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Proxy Config
  # http_proxy = "http://corporate.proxy:3128"

  ## Compress each HTTP request payload using GZIP.
  # content_encoding = "gzip"
```

### Required parameters:

* `urls`: List of strings, this is for InfluxDB clustering
support. On each flush interval, Telegraf will randomly choose one of the urls
to write to. Each URL should start with either `http://` or `udp://`
* `database`: The name of the database to write to.


### Optional parameters:

* `write_consistency`: Write consistency (clusters only), can be: "any", "one", "quorum", "all".
* `retention_policy`:  Name of existing retention policy to write to.  Empty string writes to the default retention policy.
* `timeout`: Write timeout (for the InfluxDB client), formatted as a string. If not provided, will default to 5s. 0s means no timeout (not recommended).
* `username`: Username for influxdb
* `password`: Password for influxdb
* `user_agent`:  Set the user agent for HTTP POSTs (can be useful for log differentiation)
* `udp_payload`: Set UDP payload size, defaults to InfluxDB UDP Client default (512 bytes)
* `ssl_ca`: SSL CA
* `ssl_cert`: SSL CERT
* `ssl_key`: SSL key
* `insecure_skip_verify`: Use SSL but skip chain & host verification (default: false)
* `http_proxy`: HTTP Proxy URI
* `content_encoding`: Compress each HTTP request payload using gzip if set to: "gzip"
