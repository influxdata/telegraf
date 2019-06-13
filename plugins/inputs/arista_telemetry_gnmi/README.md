# Arista gNMI Telemetry Input Plugin

Arista gNMI telemetry is an input plugin that consumes telemetry data similar to the [GNMI specification](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md).This GRPC-based protocol can utilize TLS for authentication and encryption.

This plugin has been developed to support gNMI telemetry as produced by Arista 4.20.2.1F EOS and later.

### Configuration:

```toml
[[inputs.arista_telemetry_gnmi]]
  ## List of device addresses to collect telemetry from
  servers = ["localhost:6030"]

  ## Authentication details.
  username = "user"
  password = "pass"

  ## redial in case of failures after
  redial = "10s"

  ## enable client-side TLS and define CA to authenticate the device
  # enable_tls = true
  # tls_ca = "/etc/telegraf/ca.pem"
  # insecure_skip_verify = true

  ## define client-side TLS certificate & key to authenticate to the device
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## OpenConfig paths of the subscription.
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  ## and: https://eos.arista.com/openconfig-4-20-2-1f-release-notes/

  paths = [
   "/interfaces/interface/state/counters",
   "/acl/",   
  ]


  # Stream mode (one of: "target_defined", "sample", "on_change"), mode (one of: "once", "pull", "stream") and interval
  stream_mode = "sample"
  mode = "stream"
  sample_interval = "10s"

  ## If suppression is enabled, send updates at least every X seconds anyway
  # heartbeat_interval = "60s"
```

### Example Output
```
/interfaces/interface/state/counters,host=telegraf-01,name=Ethernet3/4/2,path=/interfaces/interface/state/counters/in-broadcast-pkts,source=10.82.100.0 /in_broadcast_pkts=43i 1558796832439786099
```


