# HTTP Output Plugin

This plugin writes to a HTTP Server using the `POST Method`.

Data collected from telegraf is sent in the Request Body.

### Configuration:

```toml
# Send telegraf metrics to HTTP Server(s)
[[outputs.http]]
  ## It requires a `url` name.
  ## Will be transmitted telegraf metrics to the HTTP Server using the below URL.
  url = "http://127.0.0.1:8080/metric"
  ## http_headers option can add a custom header to the request.
  ## The value is written as a delimiter(:).
  ## Content-Type is required http header in http plugin. 
  ## so content-type of HTTP specification (plain/text, application/json, etc...) must be filled out.
  http_headers = [ "Content-Type:application/json" ]
  ## With this HTTP status code, the http plugin checks that the HTTP request is completed normally. 
  ## As a result, any status code that is not a specified status code is considered to be an error condition and processed.
  expected_status_codes = [ 200, 204 ]
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
