# gNMI (gRPC Network Management Interface) Input Plugin

This plugin consumes telemetry data based on the [gNMI][1] Subscribe method. TLS
is supported for authentication and encryption.  This input plugin is
vendor-agnostic and is supported on any platform that supports the gNMI spec.

For Cisco devices:

It has been optimized to support gNMI telemetry as produced by Cisco IOS XR
(64-bit) version 6.5.1, Cisco NX-OS 9.3 and Cisco IOS XE 16.12 and later.

[1]: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# gNMI telemetry input plugin
[[inputs.gnmi]]
  ## Address and port of the gNMI GRPC server
  addresses = ["10.49.234.114:57777"]

  ## define credentials
  username = "cisco"
  password = "cisco"

  ## gNMI encoding requested (one of: "proto", "json", "json_ietf", "bytes")
  # encoding = "proto"

  ## redial in case of failures after
  redial = "10s"

  ## enable client-side TLS and define CA to authenticate the device
  # enable_tls = false
  # tls_ca = "/etc/telegraf/ca.pem"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true

  ## define client-side TLS certificate & key to authenticate to the device
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## gNMI subscription prefix (optional, can usually be left empty)
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  # origin = ""
  # prefix = ""
  # target = ""

  ## Define additional aliases to map telemetry encoding paths to simple measurement names
  # [inputs.gnmi.aliases]
  #   ifcounters = "openconfig:/interfaces/interface/state/counters"

  [[inputs.gnmi.subscription]]
    ## Name of the measurement that will be emitted
    name = "ifcounters"

    ## Origin and path of the subscription
    ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
    ##
    ## origin usually refers to a (YANG) data model implemented by the device
    ## and path to a specific substructure inside it that should be subscribed to (similar to an XPath)
    ## YANG models can be found e.g. here: https://github.com/YangModels/yang/tree/master/vendor/cisco/xr
    origin = "openconfig-interfaces"
    path = "/interfaces/interface/state/counters"

    # Subscription mode (one of: "target_defined", "sample", "on_change") and interval
    subscription_mode = "sample"
    sample_interval = "10s"

    ## Suppress redundant transmissions when measured values are unchanged
    # suppress_redundant = false

    ## If suppression is enabled, send updates at least every X seconds anyway
    # heartbeat_interval = "60s"

  #[[inputs.gnmi.subscription]]
    # name = "descr"
    # origin = "openconfig-interfaces"
    # path = "/interfaces/interface/state/description"
    # subscription_mode = "on_change"

    ## If tag_only is set, the subscription in question will be utilized to maintain a map of
    ## tags to apply to other measurements emitted by the plugin, by matching path keys
    ## All fields from the tag-only subscription will be applied as tags to other readings,
    ## in the format <name>_<fieldBase>.
    # tag_only = true
```

## Metrics

Each configured subscription will emit a different measurement.  Each leaf in a
GNMI SubscribeResponse Update message will produce a field reading in the
measurement. GNMI PathElement keys for leaves will attach tags to the field(s).

## Example Output

```shell
ifcounters,path=openconfig-interfaces:/interfaces/interface/state/counters,host=linux,name=MgmtEth0/RP0/CPU0/0,source=10.49.234.115,descr/description=Foo in-multicast-pkts=0i,out-multicast-pkts=0i,out-errors=0i,out-discards=0i,in-broadcast-pkts=0i,out-broadcast-pkts=0i,in-discards=0i,in-unknown-protos=0i,in-errors=0i,out-unicast-pkts=0i,in-octets=0i,out-octets=0i,last-clear="2019-05-22T16:53:21Z",in-unicast-pkts=0i 1559145777425000000
ifcounters,path=openconfig-interfaces:/interfaces/interface/state/counters,host=linux,name=GigabitEthernet0/0/0/0,source=10.49.234.115,descr/description=Bar out-multicast-pkts=0i,out-broadcast-pkts=0i,in-errors=0i,out-errors=0i,in-discards=0i,out-octets=0i,in-unknown-protos=0i,in-unicast-pkts=0i,in-octets=0i,in-multicast-pkts=0i,in-broadcast-pkts=0i,last-clear="2019-05-22T16:54:50Z",out-unicast-pkts=0i,out-discards=0i 1559145777425000000
```
