# Warp10 Output Plugin

The `warp10` output plugin writes metrics to [Warp 10][].

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Write metrics to Warp 10
[[outputs.warp10]]
  # Prefix to add to the measurement.
  prefix = "telegraf."

  # URL of the Warp 10 server
  warp_url = "http://localhost:8080"

  # Write token to access your app on warp 10
  token = "Token"

  # Warp 10 query timeout
  # timeout = "15s"

  ## Print Warp 10 error body
  # print_error_body = false

  ## Max string error size
  # max_string_error_size = 511

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Output Format

Metrics are converted and sent using the [Geo Time Series][] (GTS) input format.

The class name of the reading is produced by combining the value of the
`prefix` option, the measurement name, and the field key.  A dot (`.`)
character is used as the joining character.

The GTS form provides support for the Telegraf integer, float, boolean, and
string types directly.  Unsigned integer fields will be capped to the largest
64-bit integer (2^63-1) in case of overflow.

Timestamps are sent in microsecond precision.

[Warp 10]: https://www.warp10.io
[Geo Time Series]: https://www.warp10.io/content/03_Documentation/03_Interacting_with_Warp_10/03_Ingesting_data/02_GTS_input_format
