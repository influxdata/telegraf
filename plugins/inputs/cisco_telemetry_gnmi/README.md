# Cisco GNMI telemetry

Cisco GNMI telemetry is an input plugin that consumes telemetry data similar to the [GNMI specification](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md).
This GRPC-based protocol can utilize TLS for authentication and encryption.

This plugin has been developed to support GNMI telemetry as produced by Cisco IOS XR (64-bit) version 6.5.1 and later.


### Configuration:

This is a sample configuration for the plugin.

```toml
[[inputs.cisco_telemetry_gnmi]]
  ## Address and port of the GNMI GRPC server
  service_address = "10.49.234.114:57777"

  ## define credentials
  username = "cisco"
  password = "cisco"

  ## redial in case of failures after
  redial = "10s"

  ## enable client-side TLS and define CA to authenticate the device
  # enable_tls = true
  # tls_ca = "/etc/telegraf/ca.pem"
  # insecure_skip_verify = true

  ## define client-side TLS certificate & key to authenticate to the device
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  [[inputs.cisco_telemetry_gnmi.subscription]]
    origin = "Cisco-IOS-XR-infra-statsd-oper"
    path = "infra-statistics/interfaces/interface/latest/generic-counters"

    # Subscription mode (one of: "target_defined", "sample", "on_change") and interval
    subscription_mode = "sample"
    sample_interval = "10s"

    ## Suppress redundant transmissions when measured values are unchanged
    # suppress_redundant = false

    ## If suppression is enabled, send updates at least every X seconds anyway
    # heartbeat_interval = "60s"
```
