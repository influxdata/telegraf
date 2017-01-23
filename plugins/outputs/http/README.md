# HTTP Output Plugin

This plugin writes to a HTTP Server using the `POST Method`

It requires a `url` name.

### Configuration:

```toml
# Send telegraf metrics to HTTP Server(s)
[[outputs.graylog]]
  ## Will be transmitted telegraf metrics to the HTTP Server using the below URL.
  url = "http://127.0.0.1:8080/metric"
  ## HTTP Content-Type. Default : application/json
  content_type = "application/json"
  ## Set the number of times to retry when the status code is not 200 or an error occurs during HTTP call. Default: 3
  retry = 3
  ## Configure TLS handshake timeout. Default : 3
  tls_handshake_timeout = 3
  ## Configure response header timeout in seconds. Default : 3
  response_header_timeout = 3
  ## Configure dial timeout in seconds. Default : 3
  dial_timeout = 3
  ## Configure HTTP Keep-Alive. Default : 0
  keepalive = 0
  ## Configure HTTP expect continue timeout in seconds. Default : 0
  expect_continue_timeout = 3
  ## Configure idle connection timeout in seconds. Default : 0
  idle_conn_timeout = 3

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```
