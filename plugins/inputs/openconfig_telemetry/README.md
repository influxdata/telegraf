# OpenConfig Telemetry Input Plugin

The plugin reads OpenConfig telemetry data from listed sensors. Refer to
[openconfig.net](http://openconfig.net/) for more details.

### Configuration:

```toml
# Read metrics from MQTT topic(s)
[[inputs.openconfig_telemetry]]
  server = ["localhost:1883"]

  ## Frequency to get data in milliseconds
  sampleFrequency = 2000

  ## Sensors to subscribe for
  ## A identifier for each sensor can be provided in path by separating with space
  ## Else sensor path will be used as identifier
  sensors = [
   "/interfaces/",
   "collection /components/ /lldp",
  ]

  ## x509 Certificate to use with TLS connection. If it is not provided, an insecure
  ## channel will be opened with server
  certFile = "/path/to/x509_cert_file"

  ## Option to debug incoming protobuf encoded data
  debug = true

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Tags:

- All measurements are tagged appropriately using the identifier information
  in incoming data
