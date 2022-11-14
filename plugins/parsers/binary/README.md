# Binary Parser Plugin

The `binary` data format parser parses binary protocols into metrics using
user-specified configurations.

## Configuration

```toml
[[inputs.file]]
  files = ["example.bin"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "binary"

  ## Do not error-out if none of the filter expressions below matches.
  # allow_no_match = false

  ## Specify the endianness of the data.
  ## Available values are "be" (big-endian), "le" (little-endian) and "host",
  ## where "host" means the same endianness as the machine running Telegraf.
  # endianess = "host"

  ## Interpret input as string containing hex-encoded data.
  # hex_encoding = false

  ## Multiple parsing sections are allowed
  [[inputs.file.binary]]
    ## Optional: Metric (measurement) name to use if not extracted from the data.
    # metric_name = "my_name"

    ## Definition of the message format and the extracted data.
    ## Please note that you need to define all elements of the data in the
    ## correct order with the correct length as the data is parsed in the order
    ## given.
    ## An entry can have the following properties:
    ##  name        --  Name of the element (e.g. field or tag). Can be omitted
    ##                  for special assignments (i.e. time & measurement) or if
    ##                  entry is omitted.
    ##  type        --  Data-type of the entry. Can be "int8/16/32/64", "uint8/16/32/64",
    ##                  "float32/64", "bool" and "string".
    ##                  In case of time, this can be any of "unix" (default), "unix_ms", "unix_us",
    ##                  "unix_ns" or a valid Golang time format.
    ##  bits        --  Length in bits for this entry. If omitted, the length derived from
    ##                  the "type" property will be used. For "time" 64-bit will be used
    ##                  as default.
    ##  assignment  --  Assignment of the gathered data. Can be "measurement", "time",
    ##                  "field" or "tag". If omitted "field" is assumed.
    ##  omit        --  Omit the given data. If true, the data is skipped and not added
    ##                  to the metric. Omitted entries only need a length definition
    ##                  via "bits" or "type".
    ##  terminator  --  Terminator for dynamic-length strings. Only used for "string" type.
    ##                  Valid values are "fixed" (fixed length string given by "bits"),
    ##                  "null" (null-terminated string) or a character sequence specified
    ##                  as HEX values (e.g. "0x0D0A"). Defaults to "fixed" for strings.
    ##  timezone    --  Timezone of "time" entries. Only applies to "time" assignments.
    ##                  Can be "utc", "local" or any valid Golang timezone (e.g. "Europe/Berlin")
    entries = [
      { type = "string", assignment = "measurement", terminator: "null" },
      { name = "address", type = "uint16", assignment = "tag" },
      { name = "value",   type = "float64" },
      { type = "unix", assignment = "time" },
    ]

    ## Optional: Filter evaluated before applying the configuration.
    ## This option can be used to mange multiple configuration specific for
    ## a certain message type. If no filter is given, the configuration is applied.
    # [inputs.file.binary.filter]
    #   ## Filter message by the exact length in bytes (default: N/A).
    #   # length = 0
    #   ## Filter the message by a minimum length in bytes.
    #   ## Messages longer of of equal length will pass.
    #   # length_min = 0
    #   ## List of data parts to match.
    #   ## Only if all selected parts match, the configuration will be
    #   ## applied. The "offset" is the start of the data to match in bits,
    #   ## "bits" is the length in bits and "match" is the value to match
    #   ## against. Non-byte boundaries are supported, data is always right-aligned.
    #   selection = [
    #     { offset = 0, bits = 8, match = "0x1F" },
    #   ]
    #
    #
```

In this configuration mode, you explicitly specify the field and tags you want
to scrape out of your data.

A configuration can contain multiple `binary` subsections for e.g. the file
plugin to process the binary data multiple times. This can be useful
(together with _filters_) to handle different message types.

__Please note__: The `filter` section needs to be placed __after__ the `entries`
definitions due to TOML constraints as otherwise the entries will be assigned
to the filter section.

### General options and remarks

#### `allow_no_match` (optional)

By specifying `allow_no_match` you allow the parser to silently ignore data
that does not match _any_ given configuration filter. This can be useful if
you only want to collect a subset of the available messages.

#### `endianness` (optional)

This specifies the endianness of the data. If not specified, the parser will
fallback to the "host" endianness, assuming that the message and Telegraf
machine share the same endianness.
Alternatively, you can explicitly specify big-endian format (`"be"`) or
little-endian format (`"le"`).

#### `hex_encoding` (optional)

If `true`, the input data is interpreted as a string containing hex-encoded
data like `C0 C7 21 A9`. The value is _case insensitive_ and can handle spaces,
however prefixes like ` 0x` or `x` are _not_ allowed.

### Non-byte aligned value extraction

In both, `filter` and `entries` definitions, values can be extracted at non-byte
boundaries. You can for example extract 3-bit starting at bit-offset 8. In those
cases, the result will be masked and shifted such that the resulting byte-value
is _right_ aligned. In case your 3-bit are `101` the resulting byte value is
`0x05`.

This is especially important when specifying the `match` value in the filter
section.

### Entries definitions

The entry array specifies how to dissect the message into the measurement name,
the timestamp, tags and fields.

#### `measurement` specification

When setting the `assignment` to `"measurement"`, the extracted value
will be used as the metric name, overriding other specifications.
The `type` setting is assumed to be `"string"` and can be omitted similar
to the `name` option. See [`string` type handling](#string-type-handling)
for details and further options.

### `time` specification

When setting the `assignment` to `"time"`, the extracted value
will be used as the timestamp of the metric. By default the current
time will be used for all created metrics.

The `type` setting here contains the time-format can be set to `unix`,
`unix_ms`, `unix_us`, `unix_ns`, or an accepted
[Go "reference time"][time const]. Consult the Go [time][time parse]
package for details and additional examples on how to set the time format.
If `type` is omitted the `unix` format is assumed.

For the `unix` format and derivatives, the underlying value is assumed
to be a 64-bit integer. The `bits` setting can be used to specify other
length settings. All other time-formats assume a fixed-length `string`
value to be extracted. The length of the string is automatically
determined using the format setting in `type`.

The `timezone` setting allows to convert the extracted time to the
given value timezone. By default the time will be interpreted as `utc`.
Other valid values are `local`, i.e. the local timezone configured for
the machine, or valid timezone-specification e.g. `Europe/Berlin`.

### `tag` specification

When setting the `assignment` to `"tag"`, the extracted value
will be used as a tag. The `name` setting will be the name of the tag
and the `type` will default to `string`. When specifying other types,
the extracted value will first be interpreted as the given type and
then converted to `string`.

The `bits` setting can be used to specify the length of the data to
extract and is required for fixed-length `string` types.

### `field` specification

When setting the `assignment` to `"field"` or omitting the `assignment`
setting, the extracted value will be used as a field. The `name` setting
is used as the name of the field and the `type` as type of the field value.

The `bits` setting can be used to specify the length of the data to
extract. By default the length corresponding to `type` is used.
Please see the [string](#string-type-handling) and [bool](#bool-type-handling)
specific sections when using those types.

### `string` type handling

Strings are assumed to be fixed-length strings by default. In this case, the
`bits` setting is mandatory to specify the length of the string in _bit_.

To handle dynamic strings, the `terminator` setting can be used to specify
characters to terminate the string. The two named options, `fixed` and `null`
will specify fixed-length and null-terminated strings, respectively.
Any other setting will be interpreted as hexadecimal sequence of bytes
matching the end of the string. The termination-sequence is removed from
the result.

### `bool` type handling

By default `bool` types are assumed to be _one_ bit in length. You can
specify any other length by using the `bits` setting.
When interpreting values as booleans, any zero value will be `false`,
while any non-zero value will result in `true`.

### omitting data

Parts of the data can be omitted by setting `omit = true`. In this case,
you only need to specify the length of the chunk to omit by either using
the `type` or `bits` setting. All other options can be skipped.

### Filter definitions

Filters can be used to match the length or the content of the data against
a specified reference. See the [examples section](#examples) for details.
You can also check multiple parts of the message by specifying multiple
`section` entries for a filter. Each `section` is then matched separately.
All have to match to apply the configuration.

#### `length` and `length_min` options

Using the `length` option, the filter will check if the data to parse has
exactly the given number of _bytes_. Otherwise, the configuration will not
be applied.
Similarly, for `length_min` the data has to have _at least_ the given number
of _bytes_ to generate a match.

#### `selection` list

Selections can be used with or without length constraints to match the content
of the data. Here, the `offset` and `bits` properties will specify the start
and length of the data to check. Both values are in _bit_ allowing for non-byte
aligned value extraction. The extracted data will the be checked against the
given `match` value specified in HEX.

If multiple `selection` entries are specified _all_ of the selections must
match for the configuration to get applied.

## Examples

In the following example, we use a binary protocol with three different messages
in little-endian format

### Message A definition

```text
+--------+------+------+--------+--------+------------+--------------------+--------------------+
| ID     | type | len  | addr   | count  | failure    | value              | timestamp          |
+--------+------+------+--------+--------+------------+--------------------+--------------------+
| 0x0201 | 0x0A | 0x18 | 0x7F01 | 0x2A00 | 0x00000000 | 0x6F1283C0CA210940 | 0x10D4DF6200000000 |
+--------+------+------+--------+--------+------------+--------------------+--------------------+
```

### Message B definition

```text
+--------+------+------+------------+
| ID     | type | len  | value      |
+--------+------+------+------------+
| 0x0201 | 0x0B | 0x04 | 0xDEADC0DE |
+--------+------+------+------------+
```

### Message C definition

```text
+--------+------+------+------------+------------+--------------------+
| ID     | type | len  | value x    | value y    | timestamp          |
+--------+------+------+------------+------------+--------------------+
| 0x0201 | 0x0C | 0x10 | 0x4DF82D40 | 0x5F305C08 | 0x10D4DF6200000000 |
+--------+------+------+------------+------------+--------------------+
```

All messages consists of a 4-byte header containing the _message type_
in the 3rd byte and a message specific body. To parse those messages
you can use the following configuration

```toml
[[inputs.file]]
  files = ["messageA.bin", "messageB.bin", "messageC.bin"]
  data_format = "binary"
  endianess = "le"

  [[inputs.file.binary]]
    metric_name = "messageA"

    entries = [
      { bits = 32, omit = true },
      { name = "address", type = "uint16", assignment = "tag" },
      { name = "count",   type = "int16" },
      { name = "failure", type = "bool", bits = 32, assignment = "tag" },
      { name = "value",   type = "float64" },
      { type = "unix",    assignment = "time" },
    ]

    [inputs.file.binary.filter]
      selection = [{ offset = 16, bits = 8, match = "0x0A" }]

  [[inputs.file.binary]]
    metric_name = "messageB"

    entries = [
      { bits = 32, omit = true },
      { name = "value",   type = "uint32" },
    ]

    [inputs.file.binary.filter]
      selection = [{ offset = 16, bits = 8, match = "0x0B" }]

  [[inputs.file.binary]]
    metric_name = "messageC"

    entries = [
      { bits = 32, omit = true },
      { name = "x",   type = "float32" },
      { name = "y",   type = "float32" },
      { type = "unix",    assignment = "time" },
    ]

    [inputs.file.binary.filter]
      selection = [{ offset = 16, bits = 8, match = "0x0C" }]
```

The above configuration has one `[[inputs.file.binary]]` section per
message type and uses a filter in each of those sections to apply
the correct configuration by comparing the 3rd byte (containing
the message type). This will lead to the following output

```text
> metricA,address=383,failure=false count=42i,value=3.1415 1658835984000000000
> metricB value=3737169374i 1658847037000000000
> metricC x=2.718280076980591,y=0.0000000000000000000000000000000006626070178575745 1658835984000000000
```

where `metricB` uses the parsing time as timestamp due to missing
information in the data. The other two metrics use the timestamp
derived from the data.

[time const]:   https://golang.org/pkg/time/#pkg-constants
[time parse]:   https://golang.org/pkg/time/#Parse
