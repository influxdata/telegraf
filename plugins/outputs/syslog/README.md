# Syslog Output Plugin

The syslog output plugin sends syslog messages transmitted over
[UDP](https://tools.ietf.org/html/rfc5426) or
[TCP](https://tools.ietf.org/html/rfc6587) or
[TLS](https://tools.ietf.org/html/rfc5425), with or without the octet counting framing.

Syslog messages areformatted according to
[RFC 5424](https://tools.ietf.org/html/rfc5424).

### Configuration

```toml
## URL to connect to
# address = "tcp://127.0.0.1:8094"
# address = "tcp://example.com:http"
# address = "tcp4://127.0.0.1:8094"
# address = "tcp6://127.0.0.1:8094"
# address = "tcp6://[2001:db8::1]:8094"
# address = "udp://127.0.0.1:8094"
# address = "udp4://127.0.0.1:8094"
# address = "udp6://127.0.0.1:8094"

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

## The framing technique with which it is expected that messages are transported (default = "octet-counting").
## Whether the messages come using the octect-counting (RFC5425#section-4.3.1, RFC6587#section-3.4.1),
## or the non-transparent framing technique (RFC6587#section-3.4.2).
## Must be one of "octect-counting", "non-transparent".
# framing = "octet-counting"

## The trailer to be expected in case of non-trasparent framing (default = "LF").
## Must be one of "LF", or "NUL".
# trailer = "LF"

### SD-PARAMs settings
### A syslog message can contain multiple parameters and multiple identifiers within structured data section
### For each unrecognised metric field a SD-PARAMS can be created. 
### Example
### Configuration =>
### sdparam_separator = "_"
### default_sdid = "default@32473"
### sdids = ["foo@123", "bar@456"]
### input => xyzzy,x=y foo@123_value=42,bar@456_value2=84,something_else=1
### output (structured data only) => [foo@123 value=42][bar@456 value2=84][default@32473 something_else=1 x=y]

## SD-PARAMs separator between the sdid and field key (default = "_") 
sdparam_separator = "_"

## Default sdid used for for fields that don't contain a prefix defined in the explict sdids setting below
## If no default is specified, no SD-PARAMs will be used for unrecognised field.
#default_sdid = "default@32473"

##List of explicit prefixes to extract from fields and use as the SDID, if they match (see above example for more details):
#sdids = ["foo@123", "bar@456"]
###

# Default PRI value (RFC5424#section-6.2.1) If no metric Field with key "PRI" is defined, this default value is used.
default_priority = 0

# Default APP-NAME value (RFC5424#section-6.2.5) If no metric Field with key "APP-NAME" is defined, this default value is used.
default_appname = "Telegraf"
```