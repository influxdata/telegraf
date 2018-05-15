# HTTP Output Plugin

This plugin writes to a HTTP Server using the `POST Method`.

Data collected from telegraf is sent in the Request Body.

### Configuration:

```toml
# Send telegraf metrics to HTTP Server(s)
[[outputs.http]]
  ## It requires a url name.
  ## Will be transmitted telegraf metrics to the HTTP Server using the below URL.
  ## Note that not support the HTTPS.
  url = "http://127.0.0.1:8080/metric"
  ## Configure dial timeout in seconds. Default : 3
  timeout = 3
  ## http_headers option can add a custom header to the request.
  ## Content-Type is required http header in http plugin.
  ## so content-type of HTTP specification (plain/text, application/json, etc...) must be filled out.
  [outputs.http.headers]
    Content-Type = "plain/text"
  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```
