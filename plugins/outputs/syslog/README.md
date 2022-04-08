# Syslog Output Plugin

The syslog output plugin sends syslog messages transmitted over
[UDP](https://tools.ietf.org/html/rfc5426) or
[TCP](https://tools.ietf.org/html/rfc6587) or
[TLS](https://tools.ietf.org/html/rfc5425), with or without the octet counting
framing.

Syslog messages are formatted according to [RFC
5424](https://tools.ietf.org/html/rfc5424).

## Configuration

```toml
# Configuration for Syslog server to send metrics to
[[outputs.syslog]]
  ## URL to connect to
  ## ex: address = "tcp://127.0.0.1:8094"
  ## ex: address = "tcp4://127.0.0.1:8094"
  ## ex: address = "tcp6://127.0.0.1:8094"
  ## ex: address = "tcp6://[2001:db8::1]:8094"
  ## ex: address = "udp://127.0.0.1:8094"
  ## ex: address = "udp4://127.0.0.1:8094"
  ## ex: address = "udp6://127.0.0.1:8094"
  address = "tcp://127.0.0.1:8094"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Period between keep alive probes.
  ## Only applies to TCP sockets.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"

  ## The framing technique with which it is expected that messages are
  ## transported (default = "octet-counting").  Whether the messages come
  ## using the octect-counting (RFC5425#section-4.3.1, RFC6587#section-3.4.1),
  ## or the non-transparent framing technique (RFC6587#section-3.4.2).  Must
  ## be one of "octet-counting", "non-transparent".
  # framing = "octet-counting"

  ## The trailer to be expected in case of non-transparent framing (default = "LF").
  ## Must be one of "LF", or "NUL".
  # trailer = "LF"

  ## SD-PARAMs settings
  ## Syslog messages can contain key/value pairs within zero or more
  ## structured data sections.  For each unrecognized metric tag/field a
  ## SD-PARAMS is created.
  ##
  ## Example:
  ##   [[outputs.syslog]]
  ##     sdparam_separator = "_"
  ##     default_sdid = "default@32473"
  ##     sdids = ["foo@123", "bar@456"]
  ##
  ##   input => xyzzy,x=y foo@123_value=42,bar@456_value2=84,something_else=1
  ##   output (structured data only) => [foo@123 value=42][bar@456 value2=84][default@32473 something_else=1 x=y]

  ## SD-PARAMs separator between the sdid and tag/field key (default = "_")
  # sdparam_separator = "_"

  ## Default sdid used for tags/fields that don't contain a prefix defined in
  ## the explicit sdids setting below If no default is specified, no SD-PARAMs
  ## will be used for unrecognized field.
  # default_sdid = "default@32473"

  ## List of explicit prefixes to extract from tag/field keys and use as the
  ## SDID, if they match (see above example for more details):
  # sdids = ["foo@123", "bar@456"]

  ## Default severity value. Severity and Facility are used to calculate the
  ## message PRI value (RFC5424#section-6.2.1).  Used when no metric field
  ## with key "severity_code" is defined.  If unset, 5 (notice) is the default
  # default_severity_code = 5

  ## Default facility value. Facility and Severity are used to calculate the
  ## message PRI value (RFC5424#section-6.2.1).  Used when no metric field with
  ## key "facility_code" is defined.  If unset, 1 (user-level) is the default
  # default_facility_code = 1

  ## Default APP-NAME value (RFC5424#section-6.2.5)
  ## Used when no metric tag with key "appname" is defined.
  ## If unset, "Telegraf" is the default
  # default_appname = "Telegraf"
```

## Metric mapping

The output plugin expects syslog metrics tags and fields to match up with the
ones created in the [syslog input][].

The following table shows the metric tags, field and defaults used to format
syslog messages.

| Syslog field | Metric Tag | Metric Field | Default value |
| --- | --- | --- | --- |
| APP-NAME | appname | - | default_appname = "Telegraf" |
| TIMESTAMP | - | timestamp | Metric's own timestamp |
| VERSION | - | version | 1 |
| PRI | - | serverity_code + (8 * facility_code)| default_severity_code=5 (notice), default_facility_code=1 (user-level)|
| HOSTNAME | hostname OR source OR host | - | os.Hostname() |
| MSGID | - | msgid | Metric name |
| PROCID | - | procid | - |
| MSG | - | msg | - |

[syslog input]: /plugins/inputs/syslog#metrics
