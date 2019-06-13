# Juniper gNMI Telemetry Input Plugin

Juniper gNMI telemetry is an input plugin that consumes telemetry data similar to the [GNMI specification](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md).This GRPC-based protocol can utilize TLS for authentication and encryption.

This plugin has been developed to support gNMI telemetry as produced by Juniper devices.

### Configuration:

```toml
[[inputs.juniper_telemetry_gnmi]]
  ## List of device addresses to collect telemetry from
  servers = ["localhost:50051"]

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
  ## https://www.juniper.net/documentation/en_US/junos/topics/reference/general/junos-telemetry-interface-grpc-sensors.html  
  ## YANG models can be found e.g. here: https://github.com/YangModels/yang/tree/master/vendor/juniper

  paths = [
   "/interfaces/interface/"
  ]


  # Stream mode (one of: "target_defined", "sample", "on_change") and interval
  subscription_mode = "sample"
  sample_interval = "10s"

```

 Example Output:/interfaces/interface,host=telegraf-01,name=ae65,path=/interfaces/interface,source=192.229.214.0 __timestamp__=1560456200258i,__junos_re_stream_creation_timestamp__=1560456200252i,__junos_re_payload_get_timestamp__=1560456200252i,aggregation/state/lag_type="LACP",aggregation/state/min_links=1i,aggregation/state/lag_speed=0i,aggregation/state/member/member="xe-7/0/3:0" 1560456200253338871

