# HTTP Listener v2 Input Plugin

HTTP Listener v2 is a service input plugin that listens for metrics sent via
HTTP.  Metrics may be sent in any supported [data format][data_format].

**Note:** The plugin previously known as `http_listener` has been renamed
`influxdb_listener`.  If you would like Telegraf to act as a proxy/relay for
InfluxDB it is recommended to use [`influxdb_listener`][influxdb_listener].

### Configuration:

This is a sample configuration for the plugin.

```toml
[[inputs.http_listener_v2]]
  ## Address and port to host HTTP listener on
  service_address = ":8080"

  ## Path to listen to.
  path = "/telegraf"

  ## HTTP methods to accept.
  methods = ["POST", "PUT"]

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 536,870,912 bytes (500 mebibytes)
  max_body_size = 0

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

  ## Optional username and password to accept for HTTP basic authentication.
  ## You probably want to make sure you have TLS configured above for this.
  basic_username = "foobar"
  basic_password = "barfoo"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Metrics:

Metrics are created from the request body and are dependant on the value of `data_format`.

### Troubleshooting:

**Send Line Protocol**
```
curl -i -XPOST 'http://localhost:8080/telegraf' --data-binary 'cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000'
```

**Send JSON**
```
curl -i -XPOST 'http://localhost:8080/telegraf' --data-binary '{"value1": 42, "value2": 42}'
```

[data_format]: /docs/DATA_FORMATS_INPUT.md
[influxdb_listener]: /plugins/inputs/influxdb_listener/README.md
