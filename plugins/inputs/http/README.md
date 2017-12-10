# HTTP Input Plugin

The HTTP input plugin collects metrics from one or more HTTP(S) endpoints.  The metrics need to be formatted in one of the supported data formats.  Each data format has its own unique set of configuration options, read more about them here:
  https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md


### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage http`.

```toml
# Read formatted metrics from one or more HTTP endpoints
[[inputs.http]]
  ## One or more URLs from which to read formatted metrics
  urls = [
    "http://localhost/metrics"
  ]

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  # timeout = "5s"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"
```

### Metrics:

The metrics collected by this input plugin will depend on the configurated `data_format` and the payload returned by the HTTP endpoint(s).
