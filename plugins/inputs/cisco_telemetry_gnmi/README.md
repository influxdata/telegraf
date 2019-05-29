# Cisco GNMI telemetry

Cisco GNMI telemetry is an input plugin that consumes telemetry data similar to the [GNMI specification](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md).
This GRPC-based protocol can utilize TLS for authentication and encryption.

This plugin has been developed to support GNMI telemetry as produced by Cisco IOS XR (64-bit) version 6.5.1 and later.


### Configuration:

This is a sample configuration for the plugin.

```toml
[[inputs.cisco_telemetry_gnmi]]
  ## Address and port of the GNMI GRPC server
  address = "10.49.234.114:57777"

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

  ## GNMI encoding requested (usually one of: "proto", "json", "json_ietf")
  # encoding = "proto"

  ## GNMI subscription prefix (optional, platform dependent)
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  # origin = "oc-if"
  # prefix = "interfaces/interface"
  # target = ""


  [[inputs.cisco_telemetry_gnmi.subscription]]
    ## Origin and path of the subscription
    ## origin usually refers to a (YANG) data model implemented by the device
    ## and path to a specific substructe inside it (similar to an XPath) that should be subscribed to
    ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
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

### Example Output
```
openconfig:/interfaces/interface/state/counters,host=linux,name=MgmtEth0/RP0/CPU0/0,source=10.49.234.115 in-multicast-pkts=0i,out-multicast-pkts=0i,out-errors=0i,out-discards=0i,in-broadcast-pkts=0i,out-broadcast-pkts=0i,in-discards=0i,in-unknown-protos=0i,in-errors=0i,out-unicast-pkts=0i,in-octets=0i,out-octets=0i,last-clear="2019-05-22T16:53:21Z",in-unicast-pkts=0i 1559145777425000000
openconfig:/interfaces/interface/state/counters,host=linux,name=GigabitEthernet0/0/0/0,source=10.49.234.115 out-multicast-pkts=0i,out-broadcast-pkts=0i,in-errors=0i,out-errors=0i,in-discards=0i,out-octets=0i,in-unknown-protos=0i,in-unicast-pkts=0i,in-octets=0i,in-multicast-pkts=0i,in-broadcast-pkts=0i,last-clear="2019-05-22T16:54:50Z",out-unicast-pkts=0i,out-discards=0i 1559145777425000000
```
