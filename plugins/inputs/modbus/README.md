# Modbus Input Plugin

The Modbus plugin collects Discrete Inputs, Coils, Input Registers and Holding
Registers via Modbus TCP or Modbus RTU/ASCII.

## Configuration

```toml
# Retrieve data from MODBUS slave devices
[[inputs.modbus]]
  ## Connection Configuration
  ##
  ## The plugin supports connections to PLCs via MODBUS/TCP, RTU over TCP, ASCII over TCP or
  ## via serial line communication in binary (RTU) or readable (ASCII) encoding
  ##
  ## Device name
  name = "Device"

  ## Slave ID - addresses a MODBUS device on the bus
  ## Range: 0 - 255 [0 = broadcast; 248 - 255 = reserved]
  slave_id = 1

  ## Timeout for each request
  timeout = "1s"

  ## Maximum number of retries and the time to wait between retries
  ## when a slave-device is busy.
  # busy_retries = 0
  # busy_retries_wait = "100ms"

  # TCP - connect via Modbus/TCP
  controller = "tcp://localhost:502"

  ## Serial (RS485; RS232)
  # controller = "file:///dev/ttyUSB0"
  # baud_rate = 9600
  # data_bits = 8
  # parity = "N"
  # stop_bits = 1

  ## For Modbus over TCP you can choose between "TCP", "RTUoverTCP" and "ASCIIoverTCP"
  ## default behaviour is "TCP" if the controller is TCP
  ## For Serial you can choose between "RTU" and "ASCII"
  # transmission_mode = "RTU"

  ## Trace the connection to the modbus device as debug messages
  ## Note: You have to enable telegraf's debug mode to see those messages!
  # debug_connection = false

  ## Define the configuration schema
  ##  |---register -- define fields per register type in the original style (only supports one slave ID)
  ##  |---request  -- define fields on a requests base
  configuration_type = "register"

  ## --- "register" configuration style ---

  ## Measurements
  ##

  ## Digital Variables, Discrete Inputs and Coils
  ## measurement - the (optional) measurement name, defaults to "modbus"
  ## name        - the variable name
  ## address     - variable address

  discrete_inputs = [
    { name = "start",          address = [0]},
    { name = "stop",           address = [1]},
    { name = "reset",          address = [2]},
    { name = "emergency_stop", address = [3]},
  ]
  coils = [
    { name = "motor1_run",     address = [0]},
    { name = "motor1_jog",     address = [1]},
    { name = "motor1_stop",    address = [2]},
  ]

  ## Analog Variables, Input Registers and Holding Registers
  ## measurement - the (optional) measurement name, defaults to "modbus"
  ## name        - the variable name
  ## byte_order  - the ordering of bytes
  ##  |---AB, ABCD   - Big Endian
  ##  |---BA, DCBA   - Little Endian
  ##  |---BADC       - Mid-Big Endian
  ##  |---CDAB       - Mid-Little Endian
  ## data_type   - INT16, UINT16, INT32, UINT32, INT64, UINT64,
  ##               FLOAT32-IEEE, FLOAT64-IEEE (the IEEE 754 binary representation)
  ##               FLOAT32, FIXED, UFIXED (fixed-point representation on input)
  ## scale       - the final numeric variable representation
  ## address     - variable address

  holding_registers = [
    { name = "power_factor", byte_order = "AB",   data_type = "FIXED", scale=0.01,  address = [8]},
    { name = "voltage",      byte_order = "AB",   data_type = "FIXED", scale=0.1,   address = [0]},
    { name = "energy",       byte_order = "ABCD", data_type = "FIXED", scale=0.001, address = [5,6]},
    { name = "current",      byte_order = "ABCD", data_type = "FIXED", scale=0.001, address = [1,2]},
    { name = "frequency",    byte_order = "AB",   data_type = "UFIXED", scale=0.1,  address = [7]},
    { name = "power",        byte_order = "ABCD", data_type = "UFIXED", scale=0.1,  address = [3,4]},
  ]
  input_registers = [
    { name = "tank_level",   byte_order = "AB",   data_type = "INT16",   scale=1.0,     address = [0]},
    { name = "tank_ph",      byte_order = "AB",   data_type = "INT16",   scale=1.0,     address = [1]},
    { name = "pump1_speed",  byte_order = "ABCD", data_type = "INT32",   scale=1.0,     address = [3,4]},
  ]

  ## --- "request" configuration style ---

  ## Per request definition
  ##

  ## Define a request sent to the device
  ## Multiple of those requests can be defined. Data will be collated into metrics at the end of data collection.
  [[inputs.modbus.request]]
    ## ID of the modbus slave device to query.
    ## If you need to query multiple slave-devices, create several "request" definitions.
    slave_id = 1

    ## Byte order of the data.
    ##  |---ABCD -- Big Endian (Motorola)
    ##  |---DCBA -- Little Endian (Intel)
    ##  |---BADC -- Big Endian with byte swap
    ##  |---CDAB -- Little Endian with byte swap
    byte_order = "ABCD"

    ## Type of the register for the request
    ## Can be "coil", "discrete", "holding" or "input"
    register = "coil"

    ## Name of the measurement.
    ## Can be overriden by the individual field definitions. Defaults to "modbus"
    # measurement = "modbus"

    ## Field definitions
    ## Analog Variables, Input Registers and Holding Registers
    ## address        - address of the register to query. For coil and discrete inputs this is the bit address.
    ## name *1        - field name
    ## type *1,2      - type of the modbus field, can be INT16, UINT16, INT32, UINT32, INT64, UINT64 and
    ##                  FLOAT32, FLOAT64 (IEEE 754 binary representation)
    ## scale *1,2     - (optional) factor to scale the variable with
    ## output *1,2    - (optional) type of resulting field, can be INT64, UINT64 or FLOAT64. Defaults to FLOAT64 if
    ##                  "scale" is provided and to the input "type" class otherwise (i.e. INT* -> INT64, etc).
    ## measurement *1 - (optional) measurement name, defaults to the setting of the request
    ## omit           - (optional) omit this field. Useful to leave out single values when querying many registers
    ##                  with a single request. Defaults to "false".
    ##
    ## *1: Those fields are ignored if field is omitted ("omit"=true)
    ##
    ## *2: Thise fields are ignored for both "coil" and "discrete"-input type of registers. For those register types
    ##     the fields are output as zero or one in UINT64 format by default.

    ## Coil / discrete input example
    fields = [
      { address=0, name="motor1_run"},
      { address=1, name="jog", measurement="motor"},
      { address=2, name="motor1_stop", omit=true},
      { address=3, name="motor1_overheating"},
    ]

    [[inputs.modbus.request.tags]]
      machine = "impresser"
      location = "main building"

  [[inputs.modbus.request]]
    ## Holding example
    ## All of those examples will result in FLOAT64 field outputs
    slave_id = 1
    byte_order = "DCBA"
    register = "holding"
    fields = [
      { address=0, name="voltage",      type="INT16",   scale=0.1   },
      { address=1, name="current",      type="INT32",   scale=0.001 },
      { address=3, name="power",        type="UINT32",  omit=true   },
      { address=5, name="energy",       type="FLOAT32", scale=0.001, measurement="W" },
      { address=7, name="frequency",    type="UINT32",  scale=0.1   },
      { address=8, name="power_factor", type="INT64",   scale=0.01  },
    ]

    [[inputs.modbus.request.tags]]
      machine = "impresser"
      location = "main building"

  [[inputs.modbus.request]]
    ## Input example with type conversions
    slave_id = 1
    byte_order = "ABCD"
    register = "input"
    fields = [
      { address=0, name="rpm",         type="INT16"                   },  # will result in INT64 field
      { address=1, name="temperature", type="INT16", scale=0.1        },  # will result in FLOAT64 field
      { address=2, name="force",       type="INT32", output="FLOAT64" },  # will result in FLOAT64 field
      { address=4, name="hours",       type="UINT32"                  },  # will result in UIN64 field
    ]

    [[inputs.modbus.request.tags]]
      machine = "impresser"
      location = "main building"

  ## Enable workarounds required by some devices to work correctly
  # [inputs.modbus.workarounds]
    ## Pause between read requests sent to the device. This might be necessary for (slow) serial devices.
    # pause_between_requests = "0ms"
    ## Close the connection after every gather cycle. Usually the plugin closes the connection after a certain
    ## idle-timeout, however, if you query a device with limited simultaneous connectivity (e.g. serial devices)
    ## from multiple instances you might want to only stay connected during gather and disconnect afterwards.
    # close_connection_after_gather = false
```

## Notes

You can debug Modbus connection issues by enabling `debug_connection`. To see those debug messages, Telegraf has to be started with debugging enabled (i.e. with the `--debug` option). Please be aware that connection tracing will produce a lot of messages and should **NOT** be used in production environments.

Please use `pause_between_requests` with care. Ensure the total gather time, including the pause(s), does not exceed the configured collection interval. Note that pauses add up if multiple requests are sent!

## Configuration styles

The modbus plugin supports multiple configuration styles that can be set using the `configuration_type` setting. The different styles are described below. Please note that styles cannot be mixed, i.e. only the settings belonging to the configured `configuration_type` are used for constructing _modbus_ requests and creation of metrics.

Directly jump to the styles:

- [original / register plugin style](#register-configuration-style)
- [per-request style](#request-configuration-style)

---

### `register` configuration style

This is the original style used by this plugin. It allows a per-register configuration for a single slave-device.

#### Metrics

Metrics are custom and configured using the `discrete_inputs`, `coils`,
`holding_register` and `input_registers` options.

#### Usage of `data_type`

The field `data_type` defines the representation of the data value on input from the modbus registers.
The input values are then converted from the given `data_type` to a type that is apropriate when
sending the value to the output plugin. These output types are usually one of string,
integer or floating-point-number. The size of the output type is assumed to be large enough
for all supported input types. The mapping from the input type to the output type is fixed
and cannot be configured.

##### Integers: `INT16`, `UINT16`, `INT32`, `UINT32`, `INT64`, `UINT64`

These types are used for integer input values. Select the one that matches your modbus data source.

##### Floating Point: `FLOAT32-IEEE`, `FLOAT64-IEEE`

Use these types if your modbus registers contain a value that is encoded in this format. These types
always include the sign, therefore no variant exists.

##### Fixed Point: `FIXED`, `UFIXED` (`FLOAT32`)

These types are handled as an integer type on input, but are converted to floating point representation
for further processing (e.g. scaling). Use one of these types when the input value is a decimal fixed point
representation of a non-integer value.

Select the type `UFIXED` when the input type is declared to hold unsigned integer values, which cannot
be negative. The documentation of your modbus device should indicate this by a term like
'uint16 containing fixed-point representation with N decimal places'.

Select the type `FIXED` when the input type is declared to hold signed integer values. Your documentation
of the modbus device should indicate this with a term like 'int32 containing fixed-point representation
with N decimal places'.

(FLOAT32 is deprecated and should not be used. UFIXED provides the same conversion from unsigned values).

---

### `request` configuration style

This sytle can be used to specify the modbus requests directly. It enables specifying multiple `[[inputs.modbus.request]]` sections including multiple slave-devices. This way, _modbus_ gateway devices can be queried. Please note that _requests_ might be split for non-consecutive addresses. If you want to avoid this behavior please add _fields_ with the `omit` flag set filling the gaps between addresses.

#### Slave device

You can use the `slave_id` setting to specify the ID of the slave device to query. It should be specified for each request, otherwise it defaults to zero. Please note, only one `slave_id` can be specified per request.

#### Byte order of the register

The `byte_order` setting specifies the byte and word-order of the registers. It can be set to `ABCD` for _big endian (Motorola)_ or `DCBA` for _little endian (Intel)_ format as well as `BADC` and `CDAB` for _big endian_ or _little endian_ with _byte swap_.

#### Register type

The `register` setting specifies the modbus register-set to query and can be set to `coil`, `discrete`, `holding` or `input`.

#### Per-request measurement setting

You can specify the name of the measurement for the following field definitions using the `measurement` setting. If the setting is omitted `modbus` is used. Furthermore, the measurement value can be overridden by each field individually.

#### Field definitions

Each `request` can contain a list of fields to collect from the modbus device.

##### address

A field is identified by an `address` that reflects the modbus register address. You can usually find the address values for the different datapoints in the datasheet of your modbus device. This is a mandatory setting.

For _coil_ and _discrete input_ registers this setting specifies the __bit__ containing the value of the field.

##### name

Using the `name` setting you can specify the field-name in the metric as output by the plugin. This setting is ignored if the field's `omit` is set to `true` and can be omitted in this case.

__Please note:__ There cannot be multiple fields with the same `name` in one metric identified by `measurement`, `slave_id` and `register`.

##### register datatype

The `register` setting specifies the datatype of the modbus register and can be set to `INT16`, `UINT16`, `INT32`, `UINT32`, `INT64` or `UINT64` for integer types or `FLOAT32` and `FLOAT64` for IEEE 754 binary representations of floating point values. Usually the datatype of the register is listed in the datasheet of your modbus device in relation to the `address` described above.

 This setting is ignored if the field's `omit` is set to `true` or if the `register` type is a bit-type (`coil` or `discrete`) and can be omitted in these cases.

##### scaling

You can use the `scale` setting to scale the register values, e.g. if the register contains a fix-point values in `UINT32` format with two decimal places for example. To convert the read register value to the actual value you can set the `scale=0.01`. The scale is used as a factor e.g. `field_value * scale`.

This setting is ignored if the field's `omit` is set to `true` or if the `register` type is a bit-type (`coil` or `discrete`) and can be omitted in these cases.

__Please note:__ The resulting field-type will be set to `FLOAT64` if no output format is specified.

##### output datatype

Using the `output` setting you can explicitly specify the output field-datatype. The `output` type can be `INT64`, `UINT64` or `FLOAT64`. If not set explicitly, the output type is guessed as follows: If `scale` is set to a non-zero value, the output type is `FLOAT64`. Otherwise, the output type corresponds to the register datatype _class_, i.e. `INT*` will result in `INT64`, `UINT*` in `UINT64` and `FLOAT*` in `FLOAT64`.

This setting is ignored if the field's `omit` is set to `true` or if the `register` type is a bit-type (`coil` or `discrete`) and can be omitted in these cases. For `coil` and `discrete` registers the field-value is output as zero or one in `UINT16` format.

#### per-field measurement setting

The `measurement` setting can be used to override the measurement name on a per-field basis. This might be useful if you want to split the fields in one request to multiple measurements. If not specified, the value specified in the [`request` section](#per-request-measurement-setting) or, if also omitted, `modbus` is used.

This setting is ignored if the field's `omit` is set to `true` and can be omitted in this case.

#### omitting a field

When specifying `omit=true`, the corresponding field will be ignored when collecting the metric but is taken into account when constructing the modbus requests. This way, you can fill "holes" in the addresses to construct consecutive address ranges resulting in a single request. Using a single modbus request can be beneficial as the values are all collected at the same point in time.

#### Tags definitions

Each `request` can be accompanied by tags valid for this request.
__Please note:__ These tags take precedence over predefined tags such as `name`, `type` or `slave_id`.

---

## Troubleshooting

### Strange data

Modbus documentation is often a mess. People confuse memory-address (starts at one) and register address (starts at zero) or are unsure about the word-order used. Furthermore, there are some non-standard implementations that also swap the bytes within the register word (16-bit).

If you get an error or don't get the expected values from your device, you can try the following steps (assuming a 32-bit value).

If you are using a serial device and get a `permission denied` error, check the permissions of your serial device and change them accordingly.

In case you get an `exception '2' (illegal data address)` error you might try to offset your `address` entries by minus one as it is very likely that there is confusion between memory and register addresses.

If you see strange values, the `byte_order` might be wrong. You can either probe all combinations (`ABCD`, `CDBA`, `BADC` or `DCBA`) or set `byte_order="ABCD" data_type="UINT32"` and use the resulting value(s) in an online converter like [this](https://www.scadacore.com/tools/programming-calculators/online-hex-converter/). This especially makes sense if you don't want to mess with the device, deal with 64-bit values and/or don't know the `data_type` of your register (e.g. fix-point floating values vs. IEEE floating point).

If your data still looks corrupted, please post your configuration, error message and/or the output of `byte_order="ABCD" data_type="UINT32"` to one of the telegraf support channels (forum, slack or as an issue).
If nothing helps, please post your configuration, error message and/or the output of `byte_order="ABCD" data_type="UINT32"` to one of the telegraf support channels (forum, slack or as an issue).

### Workarounds

Some Modbus devices need special read characteristics when reading data and will fail otherwise. For example, some serial devices need a pause between register read requests. Others might only support a limited number of simultaneously connected devices, like serial devices or some ModbusTCP devices. In case you need to access those devices in parallel you might want to disconnect immediately after the plugin finishes reading.

To enable this plugin to also handle those "special" devices, there is the `workarounds` configuration option. In case your documentation states certain read requirements or you get read timeouts or other read errors, you might want to try one or more workaround options.
If you find that other/more workarounds are required for your device, please let us know.

In case your device needs a workaround that is not yet implemented, please open an issue or submit a pull-request.

## Example Output

```sh
$ ./telegraf -config telegraf.conf -input-filter modbus -test
modbus.InputRegisters,host=orangepizero Current=0,Energy=0,Frecuency=60,Power=0,PowerFactor=0,Voltage=123.9000015258789 1554079521000000000
```
