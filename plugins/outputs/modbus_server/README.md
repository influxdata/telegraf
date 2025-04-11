# Modbus Server Output Plugin

The `modbus_server` output plugin sends data to a Modbus server.
This plugin supports various data types and allows for flexible
configuration of metrics and their corresponding Modbus registers.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml
[[outputs.modbus_server]]
[[outputs.allseas_modbus_server]]
  ## The address of the Modbus server (e.g., "tcp://localhost:502").
  server_address = "tcp://localhost:502"

  ## Byte order of the Modbus registers. Supported values are "ABCD", "BADC", "CDAB", "DCBA".
  byte_order = "ABCD"

  ## Timeout for Modbus requests (sec).
  timeout = 30

  ## Maximum number of concurrent clients.
  max_clients = 5

  ## Metrics to send to the Modbus server.
  [[outputs.modbus_server.metrics]]
    measurement = "metric_name"
    tags = { tag1 = "value1", tag2 = "value2" }

    [[outputs.modbus_server.metrics.fields]]
  [[outputs.allseas_modbus_server.metrics]]
    measurement = "metric_name"
    tags = { tag1 = "value1", tag2 = "value2" }

    [[outputs.allseas_modbus_server.metrics.fields]]
      register = "coil"
      address = 1
      name = "field1"
      type = "BIT"

    [[outputs.modbus_server.metrics.fields]]
    [[outputs.allseas_modbus_server.metrics.fields]]
      register = "register"
      address = 2
      name = "field2"
      type = "UINT16"
```

## Metric Schema

The metrics section defines the metrics to be collected from the Modbus server.
Each metric can have multiple fields, each corresponding to a Modbus register.

## Fields

- register: The type of Modbus register (e.g., "coil", "register").
- address: The address of the Modbus register.
- name: The name of the field.
- value: The value of the field.
- type: The data type of the field. Supported types are:
  - `BIT`
  - `UINT16`
  - `FLOAT32`
  - `INT32`
  - `UINT32`
  - `INT64`
  - `UINT64`
  - `FLOAT64`
  - `INT8L`
  - `INT8H`
  - `UINT8L`
  - `UINT8H`
  - `FLOAT16`
    - `STRING`

Only the tags, field name and values are part of the output metric.
The address and type are used to read the data from the Modbus server.
