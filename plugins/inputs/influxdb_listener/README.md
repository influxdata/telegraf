# InfluxDB Listener Input Plugin

InfluxDB Listener is a service input plugin that listens for requests sent
according to the [InfluxDB HTTP API][influxdb_http_api].  The intent of the
plugin is to allow Telegraf to serve as a proxy/router for the `/write`
endpoint of the InfluxDB HTTP API.

**Note:** This plugin was previously known as `http_listener`.  If you wish to
send general metrics via HTTP it is recommended to use the
[`http_listener_v2`][http_listener_v2] instead.

The `/write` endpoint supports the `precision` query parameter and can be set
to one of `ns`, `u`, `ms`, `s`, `m`, `h`.  All other parameters are ignored and
defer to the output plugins configuration.

When chaining Telegraf instances using this plugin, CREATE DATABASE requests
receive a 200 OK response with message body `{"results":[]}` but they are not
relayed. The output configuration of the Telegraf instance which ultimately
submits data to InfluxDB determines the destination database.

### Configuration:

```toml
[[inputs.influxdb_listener]]
  ## Address and port to host HTTP listener on
  service_address = ":8186"

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 536,870,912 bytes (500 mebibytes)
  max_body_size = 0

  ## Maximum line size allowed to be sent in bytes.
  ## 0 means to use the default of 65536 bytes (64 kibibytes)
  max_line_size = 0

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

  ## Optional username and password to accept for HTTP basic authentication.
  ## You probably want to make sure you have TLS configured above for this.
  # basic_username = "foobar"
  # basic_password = "barfoo"
```

### Metrics:

Metrics are created from InfluxDB Line Protocol in the request body.

### Troubleshooting:

**Example Query:**
```
curl -i -XPOST 'http://localhost:8186/write' --data-binary 'cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000'
```

[influxdb_http_api]: https://docs.influxdata.com/influxdb/latest/guides/writing_data/
[http_listener_v2]: /plugins/inputs/http_listener_v2/README.md
