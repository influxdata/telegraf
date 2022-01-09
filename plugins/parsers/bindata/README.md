# BinData

The "BinData" parser translates binary records consisting of multiple fields into Telegraf metrics. It supports:

- Little- and Big-Endian encoding
- bool, int8/uint8, int16/uint16, int32/uint32, int64/uint64, float32/float64 field types
- UTF-8 and ASCII-encoded strings
- unix, unix_ms, unix_us and unix_ns timestamp

### Configuration

```toml
[[inputs.mqtt_consumer]]
  name_override = "drone_status"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "bindata"

  ## Numeric fields endiannes, "be" or "le", default "be"
  # bindata_endiannes = "be"

  ## Timestamp format - "unix", "unix_ms", "unix_us", "unix_ns", default "unix"
  # bindata_time_format = "unix"

  ## String encoding, default "UTF-8"
  bindata_string_encoding = "UTF-8"

  ## Binary data descriptor
  ## Fields are described by:
  ## - name - field name, use Golang naming conventions
  ## - type - field type, use Golang type names
  ## - size - size in bytes, obligatory for fields with type "string" and "padding", ignored in numeric and bool fields
  ## Field with case incensitive name "time" has special meaning - it is used as metric time and must be of type
  ## - int32 for bindata_time_format = "unix", or of type
  ## - int64 for bindata_time_format = "unix_ms", "unix_us", "unix_ns".
  ## Use padding when auto-generated metric time is preferred.
  bindata_fields = [
    {name="Version",type="uint16"},
    {name="Time",type="int32"},
    {name="Latitude",type="float64"},
    {name="Longitude",type="float64"},
    {name="Altitude",type="float32"},
    {name="Heading",type="float32"},
    {name="Elevation",type="float32"},
    {name="Bank",type="float32"},
    {name="GroundSpeed",type="float32"},
    {name="AirSpeed",type="float32"},
    {name="None",type="padding", size=16},
    {name="Status",type="string",size=7},
    {name="StatusOK",type="bool"},
  ]
```
