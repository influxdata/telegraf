# Graylog Output Plugin

This plugin writes to a Graylog instance using the "[GELF][]" format.

[GELF]: https://docs.graylog.org/en/3.1/pages/gelf.html#gelf-payload-specification

## GELF Fields

The [GELF spec][] spec defines a number of specific fields in a GELF payload.
These fields may have specific requirements set by the spec and users of the
Graylog plugin need to follow these requirements or metrics may be rejected due
to invalid data.

For example, the timestamp field defined in the GELF spec, is required to be a
UNIX timestamp. This output plugin will not modify or check the timestamp field
if one is present and send it as-is to Graylog. If the field is absent then
Telegraf will set the timestamp to the current time.

Any field not defined by the spec will have an underscore (e.g. `_`) prefixed to
the field name.

[GELF spec]: https://docs.graylog.org/docs/gelf#gelf-payload-specification

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send telegraf metrics to graylog
[[outputs.graylog]]
  ## Endpoints for your graylog instances.
  servers = ["udp://127.0.0.1:12201"]

  ## Connection timeout.
  # timeout = "5s"

  ## The field to use as the GELF short_message, if unset the static string
  ## "telegraf" will be used.
  ##   example: short_message_field = "message"
  # short_message_field = ""

  ## According to GELF payload specification, additional fields names must be prefixed
  ## with an underscore. Previous versions did not prefix custom field 'name' with underscore.
  ## Set to true for backward compatibility.
  # name_field_no_prefix = false

  ## Connection retry options
  ## Attempt to connect to the enpoints if the initial connection fails.
  ## If 'false', Telegraf will give up after 3 connection attempt and will
  ## exit with an error. If set to 'true', the plugin will retry to connect
  ## to the unconnected endpoints infinitely.
  # connection_retry = false
  ## Time to wait between connection retry attempts.
  # connection_retry_wait_time = "15s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

Server endpoint may be specified without UDP or TCP scheme
(eg. "127.0.0.1:12201").  In such case, UDP protocol is assumed. TLS config is
ignored for UDP endpoints.
