# Generic HTTP listener service input plugin

> NOTE: This is a new version of HTTP listener plugin.
> This plugin supports all [data formats](/docs/DATA_FORMATS_INPUT.md) while the old [http_listener](/plugins/inputs/http_listener)
> only accepts data in InfluxDB line-protocol only

The HTTP listener is a service input plugin that listens for messages sent via HTTP POST.

Enable TLS by specifying the file names of a service TLS certificate and key.

Enable mutually authenticated TLS and authorize client connections by signing certificate authority by including a list of allowed CA certificate file names in ````tls_allowed_cacerts````.

Enable basic HTTP authentication of clients by specifying a username and password to check for. These credentials will be received from the client _as plain text_ if TLS is not configured.

**Example:**
```
curl -i -XPOST 'http://localhost:8080/write' --data-binary 'cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000'
```

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
