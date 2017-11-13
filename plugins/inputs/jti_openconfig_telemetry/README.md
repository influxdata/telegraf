# JTI OpenConfig Telemetry Input Plugin

This plugin reads Juniper Networks implementation of OpenConfig telemetry data from listed sensors using Junos Telemetry Interface. Refer to
[openconfig.net](http://openconfig.net/) for more details about OpenConfig and [Junos Telemetry Interface (JTI)](https://www.juniper.net/documentation/en_US/junos/topics/concept/junos-telemetry-interface-oveview.html).

### Configuration:

```toml
# Subscribe and receive OpenConfig Telemetry data using JTI
[[inputs.jti_openconfig_telemetry]]
  server = ["localhost:1883"]

  ## Frequency to get data in milliseconds
  sampleFrequency = 2000

  ## Sensors to subscribe for
  ## A identifier for each sensor can be provided in path by separating with space
  ## Else sensor path will be used as identifier. If a integer is provided before 
  ## sensor path, it will be used as reporting rate for that sensor instead of global
  ## reporting rate
  sensors = [
   "/interfaces/",
   "collection /components/ /lldp",
  ]

  ## Login credentials to be used with LoginCheck to authenticate session. Will try 
  ## to skip authentication if this is not provided
  username = "user"
  password = "pass"
  clientId = "telegraf"

  ## x509 Certificate to use with TLS connection. If it is not provided, an insecure
  ## channel will be opened with server
  certFile = "/path/to/x509_cert_file"

  ## Option to debug incoming protobuf encoded data
  debug = true

  ## To treat all string values as tags, set this to true
  strAsTags = false
```

### Tags:

- All measurements are tagged appropriately using the identifier information
  in incoming data
