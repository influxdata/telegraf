# OpenConfig Telemetry Input Plugin

Openconfig Telemetry data.
The plugin expects messages in the
[Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

### Configuration:

```toml
# Read metrics from MQTT topic(s)
[[inputs.openconfig_telemetry]]
  server = ["localhost:1883"]

  ## Frequency to get data in seconds
  sampleFrequency = 1

  ## Sensors to subscribe for
  ## A identifier for each sensor can be provided in path by separating with space
  ## Else sensor path will be used as identifier
  sensors = [
   "/oc/firewall/usage",
   "interfaces /oc/interfaces/",
  ]

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Tags:

- All measurements are tagged with the prefix data
