# JTI OpenConfig Telemetry Input Plugin

This plugin reads Juniper Networks implementation of OpenConfig telemetry data
from listed sensors using Junos Telemetry Interface. Refer to
[openconfig.net](http://openconfig.net/) for more details about OpenConfig and
[Junos Telemetry Interface (JTI)][1].

[1]: https://www.juniper.net/documentation/en_US/junos/topics/concept/junos-telemetry-interface-oveview.html

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Subscribe and receive OpenConfig Telemetry data using JTI
[[inputs.jti_openconfig_telemetry]]
  ## List of device addresses to collect telemetry from
  servers = ["localhost:1883"]

  ## Authentication details. Username and password are must if device expects
  ## authentication. Client ID must be unique when connecting from multiple instances
  ## of telegraf to the same device
  username = "user"
  password = "pass"
  client_id = "telegraf"

  ## Frequency to get data
  sample_frequency = "1000ms"

  ## Sensors to subscribe for
  ## A identifier for each sensor can be provided in path by separating with space
  ## Else sensor path will be used as identifier
  ## When identifier is used, we can provide a list of space separated sensors.
  ## A single subscription will be created with all these sensors and data will
  ## be saved to measurement with this identifier name
  sensors = [
   "/interfaces/",
   "collection /components/ /lldp",
  ]

  ## We allow specifying sensor group level reporting rate. To do this, specify the
  ## reporting rate in Duration at the beginning of sensor paths / collection
  ## name. For entries without reporting rate, we use configured sample frequency
  sensors = [
   "1000ms customReporting /interfaces /lldp",
   "2000ms collection /components",
   "/interfaces",
  ]

  ## Timestamp Source
  ## Set to 'collection' for time of collection, and 'data' for using the time
  ## provided by the _timestamp field.
  # timestamp_source = "collection"

  ## Optional TLS Config
  # enable_tls = false
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Delay between retry attempts of failed RPC calls or streams. Defaults to 1000ms.
  ## Failed streams/calls will not be retried if 0 is provided
  retry_delay = "1000ms"

  ## To treat all string values as tags, set this to true
  str_as_tags = false
```

## Tags

- All measurements are tagged appropriately using the identifier information
  in incoming data

## Example Output

## Metrics
