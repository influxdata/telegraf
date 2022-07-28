# IoTDB Output Plugin

[English](./README.md) | [中文](./README_ZH.md)

The IoTDB output plugin saves Telegraf metric data to an IoTDB backend.
This plugin uses Apache IoTDB client for Golang to support session connection
and data insertion.

## Apache IoTDB

Apache IoTDB (Database for Internet of Things) is an IoT native database with
high performance for data management and analysis, deployable on the edge and
the cloud. Due to its light-weight architecture, high performance and rich
feature set together with its deep integration with Apache Hadoop, Spark and
Flink, Apache IoTDB can meet the requirements of massive data storage,
high-speed data ingestion and complex data analysis in the IoT industrial
fields.

Apache IoTDB website: <https://iotdb.apache.org>
Apache IoTDB Github: <https://github.com/apache/iotdb>

## Apache IoTDB Client for Golang

There is an Apache IoTDB Client for Golang, using native Golang API to
inserting, deleting, modifying records.

Apache IoTDB Golang Client Github: <https://github.com/apache/iotdb>

## Getting started

Before using this plugin, please configure the IP address, port number,
user name, password and other information of the database server,
as well as some data type conversion, time unit and other configurations.

There is a sample configuration: [English Configuration](./sample.conf).
And there is one in Chinese:  [中文配置样例](./sample_zh.conf).
The configuration is also provided at the end of this readme.

## Please pay attention to these points

1. IoTDB (version 0.13.x or older) **DO NOT support unsigned integer**.
There are three available options of converting uint64, which are specified by
parameter `convertUint64To`.

   - `ToInt64`, default option. If an unsigned integer is greater than
   `math.MaxInt64`, save it as `int64`; else save `math.MaxInt64`
   (9223372036854775807).
   - `ForceToInt64`, force converting an unsigned integer to a`int64`,no mater
   what the value it is. This option may lead to exception if the value is
   greater than `int64`.
   - `Text`force converting an unsigned integer to a string, no mater what the
   value it is.

2. IoTDB supports a variety of time precision, but no matter what precision,
timestamp is stored in `Int64`, so users need to specify the unit of timestamp.
Default unit is `nanosecond`.

3. Till now, IoTDB can not support Tag indexing well. To see current process
method, please refer to [InfluxDB-Protocol Adapter](
   https://iotdb.apache.org/UserGuide/Master/API/InfluxDB-Protocol.html).
There are two available options of converting tags, which are specified by
parameter `treateTagsAs`:

   - `Measurements`. Treat Tags as measurements. For each Key:Value in Tag,
   convert them into Measurement, Value, DataType, which are supported in IoTDB.
   - `DeviceID_subtree`, default option. Treat Tags as part of device id. Tags
   is subtree of 'Name'.

   For example, there is a metric:

   `Name="root.sg.device", Tags={tag1="private", tag2="working"}, Fields={s1=100, s2="hello"}`

   - `Measurements`, result: `root.sg.device, s1=100, s2="hello", tag1="private", tag2="working"`
   - `DeviceID_subtree`, result: `root.sg.device.private.working, s1=100, s2="hello"`

## Testing

**Please prepare running database before testing**.
Target address is`localhost:6667` by default, which can be edit in
`iotdb_test.go`. `test_host` is the target ip address of database server.

Testing contains: network connection, error correction, datatype conversion,
data writing.

## Configuration

```toml @sample.conf
# Save metrics to an IoTDB Database
[[outputs.iotdb]]
  ## Configuration of IoTDB server
  ## host is the IP address of the IoTDB server
  ## port is a string descirbing the port of the IoTDB service, default port is 6667
  host = "127.0.0.1"
  port = "6667"

  ## Configuration of authentication
  ## The defualt user is 'root', and the defualt password is also 'root'
  user = "root"
  password = "root"

  ## Configuration of session
  ## - timeout is in milliseconds(ms). This is the timeout for calling 'Session.Open'
  ##   The value type of 'timeout' should be int. 0 means no timeout. 
  ##   Negative values will be treated as 0.
  timeout = 5000

  ## Configuration of type conversion for 64-bits unsigned int
  ## IoTDB DO NOT support unsigned int (version 13.x). 
  ## For uint32, this plugin will convert it into int64.
  ## But if the specific type of an unsigned int is uint64, overflow may take place. 
  ## So user should choose an available option of converting uint64 below.
  ## 
  ## This plugin supports 3 different available conversions of UInt64: 
  ## - "ForceToInt64": no mater what the value it is, force covert it into int64.
  ## - "ToInt64"(default): if an unsigned int is less than MAXINT64, covert it into int64; 
  ##                       else save MAXINT64 instaed. math.MaxInt64 = 9223372036854775807
  ## - "Text": no matter what value it is, convert it into a string, which is called TEXT in IoTDB.
  convertUint64To = "ToInt64"

  ## Configuration of TimeStamp
  ## TimeStamp is always saved in 64bits int. timeStampUnit specifies the unit of timestamp. 
  ## Available value:
  ## "second", "millisecond", "microsecond", "nanosecond"(defualt)
  timeStampUnit = "nanosecond"

  ## Configuration of dealing with Tags
  ## Tag is not fully supported by IoTDB, but an instead method is provided here:
  ##     https://iotdb.apache.org/zh/UserGuide/Master/API/InfluxDB-Protocol.html
  ## 
  ## This pugin provide two available methods to deal with Tags:
  ## - "Measurements": Treat Tags as measurements. For each Key:Value in Tag, convert them 
  ##                            into Measurement, Value, DataType, which are supported in IoTDB.
  ## - "DeviceID_subtree"(default): Treat Tags as part of device id. Tags is subtree of 'Name'.
  ##
  ## For Example, a metric:
  ##      Name="root.sg.device", Tags={tag1="private", tag2="working"}, Fields={s1=100, s2="hello"}
  ## - Records in "Measurements" method:
  ##      root.sg.device, s1=100, s2="hello", tag1="private", tag2="working"
  ## - Records in "DeviceID_subtree" method:
  ##      root.sg.device.private.working, s1=100, s2="hello"
  treateTagsAs = "DeviceID_subtree"

```
