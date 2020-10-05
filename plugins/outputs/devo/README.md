# Devo Output Plugin
This plugin send metrics in influx data format to Devo platform by tcp, udp and tcp+tls connection.

### Configuration:

```toml
# Configuration for Devo platform to send metrics to
[[outputs.devo]]
  ## URL to connect to
  ## address = "tcp://127.0.0.1:8094"
  ## address = "tcp://example.com:http"
  # address = "tcp://us.elb.relay.logtrust.net:443"

  ## Optional TLS Config, Required when sending directly to Devo
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Default severity value. Severity and Facility are used to calculate the
  ## message PRI value (RFC5424#section-6.2.1).  Used when no metric field
  ## with key "severity_code" is defined.  If unset, 5 (notice) is the default
  # default_severity_code = 5

  ## Default facility value. Facility and Severity are used to calculate the
  ## message PRI value (RFC5424#section-6.2.1).  Used when no metric field with
  ## key "facility_code" is defined.  If unset, 1 (user-level) is the default
  # default_facility_code = 1

  ## Period between keep alive probes.
  ## Only applies to TCP sockets.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"

  ## Content encoding for packet-based connections (i.e. UDP, unixgram).
  ## Can be set to "gzip" or to "identity" to apply no encoding.
  ##
  # content_encoding = "identity"

  ## Data format to generate.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "json"

  ## Default Devo tag value
  ## Used when no metric tag with key "devo_tag" is defined.
  ## If unset, "my.app.telegraf.default" is the default
  ## refer here for more information:
  ## https://docs.devo.com/confluence/ndt/parsers-and-collectors/about-devo-tags
  # default_tag = "my.app.telegraf.untagged"

  ## You can also manually set your hostname to identify where these metrics come from
  ## if your logs do not have identifiable information attached to them. Otherwise
  ## the plugin will try to get the hostname from your OS directly.
  # default_hostname = "unknown"
```
