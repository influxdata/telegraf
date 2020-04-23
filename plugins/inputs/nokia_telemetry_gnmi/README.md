# Nokia GNMI Telemetry

Nokia GNMI Telemetry is an input plugin that consumes telemetry data similar to the [GNMI specification](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md).
This GRPC-based protocol can utilize TLS for authentication and encryption.

This plugin has been developed to support GNMI telemetry as produced by Nokia 7750 19.5R1 and later.


### Configuration

```toml
 ## Address and port of the GNMI GRPC server
 addresses = ["192.168.113.11:57400"]

 ## username/password, the user should have grpc access rights
 username = "grpc"
 password = "Nokia4gnmi"

 ## GNMI encoding requested (one of: "json", "bytes", "json_ietf")
 # encoding = "json"

 ## redial wait time in case of failures
 redial = "10s"

 ## enable client-side TLS and define CA to authenticate the device
 # enable_tls = true
 # tls_ca = "/etc/telegraf/ca.pem"
 # insecure_skip_verify = true

 ## define client-side TLS certificate & key to authenticate to the device
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"

 ## GNMI subscription prefix (optional, can usually be left empty)
 ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
 # origin = ""
 # prefix = ""
 # target = ""

 [[inputs.nokia_telemetry_gnmi.subscription]]
  ## Name of the measurement that will be emitted
  name = "portcounters"

  ## Origin and path of the subscription
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  ##
  ## origin usually refers to a (YANG) data model implemented by the device
  ## and path to a specific substructure inside it that should be subscribed to (similar to an XPath)
  ## YANG models can be found e.g. here: https://github.com/nokia/YangModels
  # origin = ""
  path = "/state/port[port-id=*]"

  # Subscription mode (one of: "target_defined", "sample", "on_change") and interval
  subscription_mode = "target_defined"
  sample_interval = "10s"

  ## Suppress redundant transmissions when measured values are unchanged
  # suppress_redundant = false

  ## If suppression is enabled, send updates at least every X seconds anyway
  # heartbeat_interval = "60s"
```

### Example Output
```

```
