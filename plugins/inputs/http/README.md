# HTTP Input Plugin

The HTTP input plugin gathers formatted metrics from one or more HTTP(S) endpoints. 
It requires `data_format` to be specified so it can use the corresponding Parser to convert the returned payload into measurements, fields and tags.
See [DATA_FORMATS_INPUT.md](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md) for the list of supported formats.

### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage http`.

```toml
# Read formatted metrics from one or more HTTP endpoints
[[inputs.http]]
  ## One or more URLs from which to read formatted metrics
  urls = [
    "http://localhost:2015/simple.json"
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

  ## http request & header timeout
  ## defaults to 5s if not set
  timeout = "10s"

  ## Mandatory data_format
  ## See available options at https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"
```

### Metrics:

The metrics collected by this input plugin will depend on the configurated `data_format` and the payload returned by the HTTP endpoint(s).
