# Parsing Data

Telegraf has the ability to take data in a variety of formats. Telegraf requires
configuration from the user in order to correctly parse, store, and send the
original data. Telegraf does not take the raw data and maintain it internally.

Telegraf uses an internal metric representation consisting of the metric name,
tags, fields and a timestamp, very similar to [line protocol][]. This
means that data needs to be broken up into a metric name, tags, fields, and a
timestamp. While none of these options are required, they are available to
the user and might be necessary to ensure the data is represented correctly.

[line protocol]: https://docs.influxdata.com/influxdb/cloud/reference/syntax/line-protocol/

## Parsers

The first step is to determine which parser to use. Look at the list of
[parsers][] and find one that will work with the user's data. This is generally
straightforward as the data-type will only have one parser that is actually
applicable to the data.

[parsers]: https://github.com/influxdata/telegraf/tree/master/plugins/parsers

### JSON parsers

There is an exception when it comes to JSON data. Instead of a single parser,
there are three different parsers capable of reading JSON data:

* `json`: This parser is great for flat JSON data. If the JSON is more complex
  and for example, has other objects or nested arrays, then do not use this and
  look at the other two options.
* `json_v2`: The v2 parser was created out of the need to parse JSON objects. It
  can take on more advanced cases, at the cost of additional configuration.
* `xpath_json`: The xpath parser is the most capable of the three options. While
  the xpath name may imply XML data, it can parse a variety of data types using
  XPath expressions.

## Tags and fields

The next step is to look at the data and determine how the data needs to be
split up between tags and fields. Tags are generally strings or values that a
user will want to search on. While fields are the raw data values, numeric
types, etc. Generally, data is considered to be a field unless otherwise
specified as a tag.

## Timestamp

To parse a timestamp, at the very least the users needs to specify which field
has the timestamp and what the format of the timestamp is. The format can either
be a predefined Unix timestamp or parsed using a custom format based on Go
reference time.

For Unix timestamps Telegraf understands the following settings:

| Timestamp             | Timestamp Format |
|-----------------------|------------------|
| `1709572232`          | `unix`    |
| `1709572232123`       | `unix_ms` |
| `1709572232123456`    | `unix_us` |
| `1709572232123456789` | `unix_ns` |

There are some named formats available as well:

| Timestamp                             | Named Format  |
|---------------------------------------|---------------|
| `Mon Jan _2 15:04:05 2006`            | `ANSIC`       |
| `Mon Jan _2 15:04:05 MST 2006`        | `UnixDate`    |
| `Mon Jan 02 15:04:05 -0700 2006`      | `RubyDate`    |
| `02 Jan 06 15:04 MST`                 | `RFC822`      |
| `02 Jan 06 15:04 -0700`               | `RFC822Z`     |
| `Monday, 02-Jan-06 15:04:05 MST`      | `RFC850`      |
| `Mon, 02 Jan 2006 15:04:05 MST`       | `RFC1123`     |
| `Mon, 02 Jan 2006 15:04:05 -0700`     | `RFC1123Z`    |
| `2006-01-02T15:04:05Z07:00`           | `RFC3339`     |
| `2006-01-02T15:04:05.999999999Z07:00` | `RFC3339Nano` |
| `Jan _2 15:04:05`                     | `Stamp`       |
| `Jan _2 15:04:05.000`                 | `StampMilli`  |
| `Jan _2 15:04:05.000000`              | `StampMicro`  |
| `Jan _2 15:04:05.000000000`           | `StampNano`   |

If the timestamp does not conform to any of the above, then the user can specify
a custom timestamp format, in which the user must provide the timestamp in
[Go reference time][] notation. Here are a few example timestamps and their Go
reference time equivalent:

| Timestamp                     | Go reference time             |
|-------------------------------|-------------------------------|
| `2024-03-04T17:10:32`         | `2006-01-02T15:04:05` |
| `04 Mar 24 10:10 -0700`       | `02 Jan 06 15:04 -0700` |
| `2024-03-04T10:10:32Z07:00`   | `2006-01-02T15:04:05Z07:00` |
| `2024-03-04 17:10:32.123+00`  | `2006-01-02 15:04:05.999+00` |
| `2024-03-04T10:10:32.123456Z` | `2006-01-02T15:04:05.000000Z` |
| `2024-03-04T10:10:32.123456Z` | `2006-01-02T15:04:05.999999999Z` |

Note for fractional second values, the user can use either a `9` or `0`. Using a
`0` forces a certain length, but using `9`s do not.

Please note, that timezone abbreviations are ambiguous! For example `MST`, can
stand for either Mountain Standard Time (UTC-07) or Malaysia Standard Time
(UTC+08). As such, avoid abbreviated timezones if possible.

Unix timestamps use UTC, there is no concept of a timezone for a Unix timestamp.

[Go reference time]: https://pkg.go.dev/time#pkg-constants

## Examples

Below are a few basic examples to get users started.

### CSV

Given the following data:

```csv
node,temp,humidity,alarm,time
node1,32.3,23,false,2023-03-06T16:52:23Z
node2,22.6,44,false,2023-03-06T16:52:23Z
node3,17.9,56,true,2023-03-06T16:52:23Z
```

Here is corresponding parser configuration and result:

```toml
[[inputs.file]]
files = ["test.csv"]
data_format = "csv"

csv_header_row_count = 1
csv_column_names = ["node","temp","humidity","alarm","time"]
csv_tag_columns = ["node"]
csv_timestamp_column = "time"
csv_timestamp_format = "2006-01-02T15:04:05Z"
```

```text
file,node=node1 temp=32.3,humidity=23i,alarm=false 1678121543000000000
file,node=node2 temp=22.6,humidity=44i,alarm=false 1678121543000000000
file,node=node3 temp=17.9,humidity=56i,alarm=true 1678121543000000000
```

### JSON flat data

Given the following data:

```json
{ "node": "node", "temp": 32.3, "humidity": 23, "alarm": false, "time": "1709572232123456789"}
```

Here is corresponding parser configuration:

```toml
[[inputs.file]]
files = ["test.json"]
precision = "1ns"
data_format = "json"

tag_keys = ["node"]
json_time_key = "time"
json_time_format = "unix_ns"

```

```text
file,node=node temp=32.3,humidity=23 1709572232123456789
```

### JSON Objects

Given the following data:

```json
{
    "metrics": [
        { "node": "node1", "temp": 32.3, "humidity": 23, "alarm": "false", "time": "1678121543"},
        { "node": "node2", "temp": 22.6, "humidity": 44, "alarm": "false", "time": "1678121543"},
        { "node": "node3", "temp": 17.9, "humidity": 56, "alarm": "true", "time": "1678121543"}
    ]
}
```

Here is corresponding parser configuration:

```toml
[[inputs.file]]
files = ["test.json"]
data_format = "json_v2"

[[inputs.file.json_v2]]
[[inputs.file.json_v2.object]]
  path = "metrics"
  timestamp_key = "time"
  timestamp_format = "unix"
  [[inputs.file.json_v2.object.tag]]
    path = "#.node"
  [[inputs.file.json_v2.object.field]]
    path = "#.temp"
    type = "float"
  [[inputs.file.json_v2.object.field]]
    path = "#.humidity"
    type = "int"
  [[inputs.file.json_v2.object.field]]
    path = "#.alarm"
    type = "bool"
```

```text
file,node=node1 temp=32.3,humidity=23i,alarm=false 1678121543000000000
file,node=node2 temp=22.6,humidity=44i,alarm=false 1678121543000000000
file,node=node3 temp=17.9,humidity=56i,alarm=true 1678121543000000000
```

### JSON Line Protocol

Given the following data:

```json
{
  "fields": {"temp": 32.3, "humidity": 23, "alarm": false},
  "name": "measurement",
  "tags": {"node": "node1"},
  "time": "2024-03-04T10:10:32.123456Z"
}
```

Here is corresponding parser configuration:

```toml
[[inputs.file]]
files = ["test.json"]
precision = "1us"
data_format = "xpath_json"

[[inputs.file.xpath]]
  metric_name = "/name"
  field_selection = "fields/*"
  tag_selection = "tags/*"
  timestamp = "/time"
  timestamp_format = "2006-01-02T15:04:05.999999999Z"
```

```text
measurement,node=node1 alarm="false",humidity="23",temp="32.3" 1709547032123456000
```
