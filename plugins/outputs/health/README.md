# Health Output Plugin

The health plugin provides a HTTP health check resource that can be configured
to return a failure status code based on the value of a metric.

When the plugin is healthy it will return a 200 response; when unhealthy it
will return a 503 response.  The default state is healthy, one or more checks
must fail in order for the resource to enter the failed state.

## Configuration

```toml
# Configurable HTTP health check resource based on metrics
[[outputs.health]]
  ## Address and port to listen on.
  ##   ex: service_address = "http://localhost:8080"
  ##       service_address = "unix:///var/run/telegraf-health.sock"
  # service_address = "http://:8080"

  ## The maximum duration for reading the entire request.
  # read_timeout = "5s"
  ## The maximum duration for writing the entire response.
  # write_timeout = "5s"

  ## Username and password to accept for HTTP basic authentication.
  # basic_username = "user1"
  # basic_password = "secret"

  ## Allowed CA certificates for client certificates.
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## TLS server certificate and private key.
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## One or more check sub-tables should be defined, it is also recommended to
  ## use metric filtering to limit the metrics that flow into this output.
  ##
  ## When using the default buffer sizes, this example will fail when the
  ## metric buffer is half full.
  ##
  ## namepass = ["internal_write"]
  ## tagpass = { output = ["influxdb"] }
  ##
  ## [[outputs.health.compares]]
  ##   field = "buffer_size"
  ##   lt = 5000.0
  ##
  ## [[outputs.health.contains]]
  ##   field = "buffer_size"
```

### compares

The `compares` check is used to assert basic mathematical relationships.  Use
it by choosing a field key and one or more comparisons that must hold true.  If
the field is not found on a metric no comparison will be made.

Comparisons must be hold true on all metrics for the check to pass.

### contains

The `contains` check can be used to require a field key to exist on at least
one metric.

If the field is found on any metric the check passes.
