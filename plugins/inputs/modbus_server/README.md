# Modbus Server Input Plugin

The `modbus_server` input plugin collects data from a Modbus server. This plugin supports various data types and allows for flexible configuration of metrics and their corresponding Modbus registers.

## Configuration

Below is a sample configuration for the Modbus Server input plugin:

```toml
[[inputs.modbus_server]]
  server_address = "tcp://localhost:502"
  byte_order = "ABCD"
  timeout = 10
  max_clients = 5
  [[inputs.modbus_server.metrics]]
    name = "measurement1"
    fields = [
      { register = "coil", address = 0, name = "field1"},
      { register = "holding", address = 50000, name = "float_field", type = "FLOAT32" },
      { register = "holding", address = 50001, name = "bit_field0", type = "BIT", bit = 0},
      { register = "holding", address = 50001, name = "bit_field1", type = "BIT", bit = 1},
      { register = "holding", address = 50001, name = "bit_field2", type = "BIT", bit = 2},
      { register = "holding", address = 50001, name = "bit_field3", type = "BIT", bit = 3},
      { register = "holding", address = 50001, name = "bit_field4", type = "BIT", bit = 4},
      { register = "holding", address = 50001, name = "bit_field5", type = "BIT", bit = 5},
      { register = "holding", address = 50001, name = "bit_field6", type = "BIT", bit = 6},
      { register = "holding", address = 50001, name = "bit_field7", type = "BIT", bit = 7},
      { register = "holding", address = 50001, name = "bit_field8", type = "BIT", bit = 8},
      { register = "holding", address = 50001, name = "bit_field9", type = "BIT", bit = 9},
      { register = "holding", address = 50001, name = "bit_field10", type = "BIT", bit = 10},
      { register = "holding", address = 50001, name = "bit_field11", type = "BIT", bit = 11},
      { register = "holding", address = 50001, name = "bit_field12", type = "BIT", bit = 12},
      { register = "holding", address = 50001, name = "bit_field13", type = "BIT", bit = 13},
      { register = "holding", address = 50001, name = "bit_field14", type = "BIT", bit = 14},
      { register = "holding", address = 50001, name = "bit_field15", type = "BIT", bit = 15},
    ]
    [inputs.modbus_server.metrics.tags]
      tag1 = "value1"
      tag2 = "value2"
  [[inputs.modbus_server.metrics]]
    name = "measurement2"
    fields = [
      { register = "holding", address = 40000, name = "float_field", type = "FLOAT32" },
      { register = "holding", address = 40002, name = "string_field", type = "STRING", length = 10 },
    ]
    [inputs.modbus_server.metrics.tags]
      tag3 = "3"
```

## Metric Schema

The metrics section defines the metrics to be collected from the Modbus server. Each metric can have multiple fields, each corresponding to a Modbus register.

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

Only the field name and values are part of the output metric. The address and type are used to read the data from the Modbus server.

## Example

Here is an example of how to configure the plugin to collect data from a Modbus server:

```toml
[[inputs.modbus_server]]
[[inputs.allseas_modbus_server]]
  server_address = "tcp://localhost:502"
  byte_order = "ABCD"
  timeout = 30
  max_clients = 5


  [[inputs.modbus_server.metrics]]
  [[inputs.allseas_modbus_server.metrics]]
      measurement = "temperature"
      [inputs.modbus_server.metrics.tags]
        tags = { location = "server_room" }


      [[inputs.modbus_server.metrics.fields]]
      [[inputs.allseas_modbus_server.metrics.fields]]
        register = "register"
        address = 1
        name = "temp_sensor_1"
        type = "FLOAT32"


      [[inputs.modbus_server.metrics.fields]]
      [[inputs.allseas_modbus_server.metrics.fields]]
        register = "register"
        address = 2
        name = "temp_sensor_2"
        type = "FLOAT32"

  [[inputs.modbus_server.metrics]]
  [[inputs.allseas_modbus_server.metrics]]
      measurement = "humidity"
      [inputs.modbus_server.metrics.tags]
        tags = { location = "server_room" }


      [[inputs.modbus_server.metrics.fields]]
      [[inputs.allseas_modbus_server.metrics.fields]]
        register = "register"
        address = 3
        name = "humidity_sensor_1"
        type = "FLOAT32"


      [[inputs.modbus_server.metrics.fields]]
      [[inputs.allseas_modbus_server.metrics.fields]]
        register = "register"
        address = 4
        name = "humidity_sensor_2"
        type = "FLOAT32"
```
