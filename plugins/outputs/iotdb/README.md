# IoTDB Output Plugin

This output plugin saves Telegraf metrics to an Apache IoTDB backend,
supporting session connection and data insertion.

## Apache IoTDB

Apache IoTDB (Database for Internet of Things) is an IoT native database with
high performance for data management and analysis, deployable on the edge and
the cloud. Due to its light-weight architecture, high performance and rich
feature set together with its deep integration with Apache Hadoop, Spark and
Flink, Apache IoTDB can meet the requirements of massive data storage,
high-speed data ingestion and complex data analysis in the IoT industrial
fields.

For more details consult the [Apache IoTDB website](https://iotdb.apache.org)
or the [Apache IoTDB GitHub page](https://github.com/apache/iotdb).

## Getting started

Before using this plugin, please configure the IP address, port number,
user name, password and other information of the database server,
as well as some data type conversion, time unit and other configurations.

Please see the [configuration section](#Configuration) for an example
configuration.

## Metric Translation

IoTDB uses a different data format for metric data than telegraf. It is
important to note that depending on the metrics being written, the translation
may be lossy. This plugin translates to IoTDB format in the following ways:

### Unsigned Integers

IoTDB currently **DOES NOT support unsigned integer**.
There are three available options of converting uint64, which are specified by
setting `uint64_conversion`.

- `int64_clip`, default option. If an unsigned integer is greater than
`math.MaxInt64`, save it as `int64`; else save `math.MaxInt64`
(9223372036854775807).
- `int64`, force converting an unsigned integer to a`int64`,no mater
what the value it is. This option may lead to exception if the value is
greater than `int64`.
- `text`force converting an unsigned integer to a string, no mater what the
value it is.

### Time Precision

IoTDB supports a variety of time precision. You can specify which precision
you want using the `timestamp_precision` setting. Default is `nanosecond`.
Other options are `second`, `millisecond`, `microsecond`.

### Metadata (tags)

IoTDB uses a tree model for metadata while Telegraf uses a tag model.
(See [InfluxDB-Protocol Adapter](
https://iotdb.apache.org/UserGuide/Master/API/InfluxDB-Protocol.html)
There are two available options of converting tags, which are specified by
setting `convert_tags_to`:

- `fields`. Treat Tags as measurements. For each Key:Value in Tag,
convert them into Measurement, Value, DataType, which are supported in IoTDB.
- `device_id`, default option. Treat Tags as part of device id. Tags
constitute a subtree of `Name`.

For example, there is a metric:

```markdown
Name="root.sg.device", Tags={tag1="private", tag2="working"}, Fields={s1=100, s2="hello"}
```

- `fields`, result: `root.sg.device, s1=100, s2="hello", tag1="private", tag2="working"`
- `device_id`, result: `root.sg.device.private.working, s1=100, s2="hello"`

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Save metrics to an IoTDB Database
[[outputs.iotdb]]
  ## Configuration of IoTDB server connection
  host = "127.0.0.1"
  # port = "6667"

  ## Configuration of authentication
  # user = "root"
  # password = "root"

  ## Timeout to open a new session.
  ## A value of zero means no timeout.
  # timeout = "5s"

  ## Configuration of type conversion for 64-bit unsigned int
  ## IoTDB currently DOES NOT support unsigned integers (version 13.x). 
  ## 32-bit unsigned integers are safely converted into 64-bit signed integers by the plugin,
  ## however, this is not true for 64-bit values in general as overflows may occur.
  ## The following setting allows to specify the handling of 64-bit unsigned integers.
  ## Available values are:
  ##   - "int64"       --  convert to 64-bit signed integers and accept overflows
  ##   - "int64_clip"  --  convert to 64-bit signed integers and clip the values on overflow to 9,223,372,036,854,775,807
  ##   - "text"        --  convert to the string representation of the value
  # uint64_conversion = "int64_clip"

  ## Configuration of TimeStamp
  ## TimeStamp is always saved in 64bits int. timestamp_precision specifies the unit of timestamp. 
  ## Available value:
  ## "second", "millisecond", "microsecond", "nanosecond"(default)
  # timestamp_precision = "nanosecond"

  ## Handling of tags
  ## Tags are not fully supported by IoTDB. 
  ## A guide with suggestions on how to handle tags can be found here:
  ##     https://iotdb.apache.org/UserGuide/Master/API/InfluxDB-Protocol.html
  ## 
  ## Available values are:
  ##   - "fields"     --  convert tags to fields in the measurement
  ##   - "device_id"  --  attach tags to the device ID
  ##
  ## For Example, a metric named "root.sg.device" with the tags `tag1: "private"`  and  `tag2: "working"` and
  ##  fields `s1: 100`  and `s2: "hello"` will result in the following representations in IoTDB
  ##   - "fields"     --  root.sg.device, s1=100, s2="hello", tag1="private", tag2="working"
  ##   - "device_id"  --  root.sg.device.private.working, s1=100, s2="hello"
  # convert_tags_to = "device_id"
```
