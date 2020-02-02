
# Feature request

Binary Parser similar to the existing Value Parser with a possibility of parsing binary data (records) with multiple fields.

# Propsal

Implement a Binary Parser able to parse binary data (records) containing multiple fields.

At the end the Binary Parser will also support other binary data encoding protocols such as Protobuf or CBOR.

The binary data (record) configuration is shown below, where record's fields are specified by name, type, offset and optional size. Also common parameters such as endianess, binary protocol and time field format are specified:

```toml
[[inputs.mqtt_consumer]]
  name_override = "drone_status"

  ...

  data_format = "bindata"
  bindata_endiannes = "be"
  bindata_time_format = "unix"
  bindata_string_encoding = "UTF-8"
  bindata_fields = [
    {name="version",type="uint16"},
    {name="time",type="int32"},
    {name="location_latitude",type="float64"},
    {name="location_longitude",type="float64"},
    {name="location_altitude",type="float32"},
    {name="orientation_heading",type="float32"},
    {name="orientation_elevation",type="float32"},
    {name="orientation_bank",type="float32"},
    {name="speed_ground",type="float32"},
    {name="speed_air",type="float32"},
    {name="status",type="string",size=7},
  ]
```

# Use case

Parsing binary-encoded data from IoT and other domains.