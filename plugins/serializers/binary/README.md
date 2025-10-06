# Binary Serializer Plugin

The `binary` data format serializer serializes metrics into binary protocols
using user-specified configurations.

## Configuration

```toml
[[outputs.socket_writer]]
  address = "tcp://127.0.0.1:54000"
  metric_batch_size = 1

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "binary"

  ## Specify the endianness of the data.
  ## Available values are "little" (little-endian), "big" (big-endian) and "host",
  ## where "host" means the same endianness as the machine running Telegraf.
  # endianness = "host"

  ## Definition of the message format and the serialized data.
  ## Please note that you need to define all elements of the data in the
  ## correct order.
  ## An entry can have the following properties:
  ##  read_from         --  Source of the data.
  ##                        Can be "field", "tag", "time" or "name".
  ##                        If omitted "field" is assumed.
  ##  name              --  Name of the element (e.g. field or tag).
  ##                        Can be omitted for "time" and "name".
  ##  data_format       --  Target data-type of the entry. Can be "int8/16/32/64", "uint8/16/32/64",
  ##                        "float32/64", "string".
  ##                        In case of time, this can be any of "unix" (default), "unix_ms", "unix_us",
  ##                        "unix_ns".
  ##                        If original field type is different from the target type, the field will be converted
  ##                        If loss of precision is possible, warning will be logged.
  ##  string_length     --  Length of the string in bytes. Only used for "string" type.
  ##  string_terminator --  Terminator for strings. Only used for "string" type.
  ##                        Valid values are "null", "0x00", "00", "0x01", etc.
  ##                        If original string length is greater than "string_length" the string will
  ##                        be truncated to have length of the `string + terminator = string_length`.
  ##                        If original string length is smaller than "string_length" the string
  ##                        will be padded with terminator to have length of "string_length". (e.g. "abcd\0\0\0\0\0")
  ##                        Defaults to "null" for strings.
  entries = [
    { read_from = "field", name = "addr_3", data_format="int16" },
    { read_from = "field", name = "addr_2", data_format="int16" },
    { read_from = "field", name = "addr_4_5", data_format="int32" },
    { read_from = "field", name = "addr_6_7", data_format="float32" },
    { read_from = "field", name = "addr_16_20", data_format="string", string_terminator = "null", string_length = 11 },
    { read_from = "field", name = "addr_3_sc", data_format="float64" }
  ]
```

### General options and remarks

#### Value conversion

The plugin will try to convert the value of the field to the target data type.
If the conversion is not possible without precision loss value is converted and
a warning is logged.

Conversions are allowed between all supported data types.

### Examples

In the following example, we read some registers from a Modbus device and
serialize them into a binary protocol.

```toml
# Retrieve data from MODBUS slave devices
[[inputs.modbus]]
  name = "device"
  slave_id = 1
  timeout = "1s"

  controller = "tcp://127.0.0.1:5020"

  configuration_type = "register"

  holding_registers = [
    { name = "addr_2",     byte_order = "AB",   data_type="UINT16",       scale=1.0, address = [2]   },
    { name = "addr_3",     byte_order = "AB",   data_type="UINT16",       scale=1.0, address = [3]   },
    { name = "addr_4_5",   byte_order = "ABCD", data_type="UINT32",       scale=1.0, address = [4,5] },
    { name = "addr_6_7",   byte_order = "ABCD", data_type="FLOAT32-IEEE", scale=1.0, address = [6,7] },
    { name = "addr_16_20", byte_order = "ABCD", data_type="STRING",                  address = [16,17,18,19,20] },
    { name = "addr_3_sc",  byte_order = "AB",   data_type="UFIXED",       scale=0.1, address = [3]   }
  ]

[[outputs.socket_writer]]
  address = "tcp://127.0.0.1:54000"
  metric_batch_size = 1

  data_format = "binary"
  endianness = "little"
  entries = [
    { read_from = "field", name = "addr_3",   data_format="int16" },
    { read_from = "field", name = "addr_2",   data_format="int16" },
    { read_from = "field", name = "addr_4_5", data_format="int32" },
    { read_from = "field", name = "addr_6_7",  data_format="float32" },
    { read_from = "field", name = "addr_16_20", data_format="string", string_terminator = "null", string_length = 11 },
    { read_from = "field", name = "addr_3_sc",  data_format="float64" },
    { read_from = "time", data_format="int32", time_format="unix" },
    { read_from = "name", data_format="string", string_terminator = "null", string_length = 20 }
  ]
```

On the receiving side, we expect the following message structure:

```cpp
#pragma pack(push, 1)
struct test_struct
{
  short addr_3;
  short addr_2;
  int addr_4_5;
  float addr_6_7;
  char addr_16_20[11];
  double addr_3_sc;
  int time;
  char metric_name[20];
};
#pragma pack(pop)
```

Produced message:

```text
69420700296a0900c395d343415f425f435f445f455f006766666666909a407c0082656d6f646275730000000000000000000000000000
```

| addr_3 | addr_2 | addr_4_5 | addr_6_7          | addr_16_20             | addr_3_sc          | time       | metric_name                                |
|--------|--------|----------|-------------------|------------------------|--------------------|------------|--------------------------------------------|
| 0x6942 | 0700   | 296a0900 | c395d343          | 415f425f435f445f455f00 | 6766666666909a40   | 0x7c008265 | 0x6d6f646275730000000000000000000000000000 |
| 17001  | 7      | 617001   | 423.1700134277344 | A_B_C_D_E_             | 1700.1000000000001 | 1703018620 | modbus                                     |
