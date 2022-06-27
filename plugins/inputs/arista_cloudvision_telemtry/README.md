# Arista CloudVision Telemetry Input Plugin

The Arista CloudVision Telemetry Input Plugin allows for a operator to use the [gNMI gRPC interface](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md) to stream telemetry data from CloudVision portal with telegraf.  CloudVision exports every device with the serial as the [gNMI Path Target](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#2221-path-target) so when a client connects to a single address of CloudVision the device that it is interest is then distinguished based off of the path target.  Telegraf will ask CloudVision for the inventory of devices via the [resource apis](https://aristanetworks.github.io/cloudvision-apis/examples/rest/inventory/) once returned to telegraf it will then try to stream every devices paths which are specified.

A valid [service token](https://www.arista.com/en/cg-cv/cv-service-accounts) is required for each CloudVision subscription.

## Configuration
```toml @sample.conf
[[inputs.arista_cloudvision_telemtry]]
  ## CVP Address
  addresses = "10.255.35.170:8443"
  ## redial in case of failures after
  redial = "10s"

  enable_tls = false

  ## cvp service account access token generated at /cv/settings/aaa-service-accounts
  cvptoken = "-snip-"

  [[inputs.arista_cloudvision_telemtry.subscription]]
    ## Name of the measurement
    name = "System"
    origin = "openconfig"
    path = "/interfaces/interface/state/counters"
    subscription_mode = "target_defined"
```

## Example Output
```shell
/system/config/hostname,host=DC1-LEAF1A,host-id=SN-DC1-LEAF1A /system/config/hostname="DC1-LEAF1A" 1656336056235063297
/system/config/hostname,host=DC1-SPINE1,host-id=ABC12345678 /system/config/hostname="DC1-SPINE1" 1656336055710900677
/system/config/hostname,host=DC1-L2LEAF2A,host-id=SN-DC1-L2LEAF2A /system/config/hostname="DC1-L2LEAF2A" 1656336066676127908
/system/config/hostname,host=DC1-LEAF2B,host-id=SN-DC1-LEAF2B /system/config/hostname="DC1-LEAF2B" 1656336056822297422
/system/config/hostname,host=DC1-SPINE2,host-id=SN-DC1-SPINE2 /system/config/hostname="DC1-SPINE2" 1656336160090622662
/system/config/hostname,host=DC1-L2LEAF1A,host-id=SN-DC1-L2LEAF1A /system/config/hostname="DC1-L2LEAF1A" 1656336076016108686
/system/config/hostname,host=DC1-LEAF1B,host-id=SN-DC1-LEAF1B /system/config/hostname="DC1-LEAF1B" 1656336055690373742
```
