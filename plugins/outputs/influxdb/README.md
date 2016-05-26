# InfluxDB Output Plugin

This plugin writes to [InfluxDB](https://www.influxdb.com) via HTTP or UDP.

### Required parameters:

* `urls`: List of strings, this is for InfluxDB clustering
support. On each flush interval, Telegraf will randomly choose one of the urls
to write to. Each URL should start with either `http://` or `udp://`
* `database`: The name of the database to write to.


### Optional parameters:

* `retention_policy`:  Retention policy to write to.
* `precision`: Precision of writes, valid values are "ns", "us" (or "µs"), "ms", "s", "m", "h". note: using "s" precision greatly improves InfluxDB compression.
* `timeout`: Write timeout (for the InfluxDB client), formatted as a string. If not provided, will default to 5s. 0s means no timeout (not recommended).
* `username`: Username for influxdb
* `password`: Password for influxdb
* `user_agent`:  Set the user agent for HTTP POSTs (can be useful for log differentiation)
* `udp_payload`: Set UDP payload size, defaults to InfluxDB UDP Client default (512 bytes)
  ## Optional SSL Config
* `ssl_ca`: SSL CA
* `ssl_cert`: SSL CERT
* `ssl_key`: SSL key
* `insecure_skip_verify`: Use SSL but skip chain & host verification (default: false)
* `write_consistency`: Write consistency for clusters only, can be: "any", "one", "quorom", "all"
