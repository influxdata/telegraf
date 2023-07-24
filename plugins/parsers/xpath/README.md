# XPath Parser Plugin

The XPath data format parser parses different formats into metric fields using
[XPath][xpath] expressions.

For supported XPath functions check [the underlying XPath library][xpath lib].

__NOTE:__ The type of fields are specified using [XPath functions][xpath
lib]. The only exception are _integer_ fields that need to be specified in a
`fields_int` section.

## Supported data formats

| name                                         | `data_format` setting | comment |
| -------------------------------------------- | --------------------- | ------- |
| [Extensible Markup Language (XML)][xml]      | `"xml"`               |         |
| [Concise Binary Object Representation][cbor] | `"xpath_cbor"`        | [see additional notes](#concise-binary-object-representation-notes)|
| [JSON][json]                                 | `"xpath_json"`        |         |
| [MessagePack][msgpack]                       | `"xpath_msgpack"`     |         |
| [Protocol-buffers][protobuf]                 | `"xpath_protobuf"`    | [see additional parameters](#protocol-buffers-additional-settings)|

### Protocol-buffers additional settings

For using the protocol-buffer format you need to specify additional
(_mandatory_) properties for the parser. Those options are described here.

#### `xpath_protobuf_file` (mandatory)

Use this option to specify the name of the protocol-buffer definition file
(`.proto`).

#### `xpath_protobuf_type` (mandatory)

This option contains the top-level message file to use for deserializing the
data to be parsed. Usually, this is constructed from the `package` name in the
protocol-buffer definition file and the `message` name as `<package
name>.<message name>`.

#### `xpath_protobuf_import_paths` (optional)

In case you import other protocol-buffer definitions within your `.proto` file
(i.e. you use the `import` statement) you can use this option to specify paths
to search for the imported definition file(s). By default the imports are only
searched in `.` which is the current-working-directory, i.e. usually the
directory you are in when starting telegraf.

Imagine you do have multiple protocol-buffer definitions (e.g. `A.proto`,
`B.proto` and `C.proto`) in a directory (e.g. `/data/my_proto_files`) where your
top-level file (e.g. `A.proto`) imports at least one other definition

```protobuf
syntax = "proto3";

package foo;

import "B.proto";

message Measurement {
    ...
}
```

You should use the following setting

```toml
[[inputs.file]]
  files = ["example.dat"]

  data_format = "xpath_protobuf"
  xpath_protobuf_file = "A.proto"
  xpath_protobuf_type = "foo.Measurement"
  xpath_protobuf_import_paths = [".", "/data/my_proto_files"]

  ...
```

#### `xpath_protobuf_skip_bytes` (optional)

This option allows to skip a number of bytes before trying to parse
the protocol-buffer message. This is useful in cases where the raw data
has a header e.g. for the message length or in case of GRPC messages.

This is a list of known headers and the corresponding values for
`xpath_protobuf_skip_bytes`

| name                                    | setting | comment |
| --------------------------------------- | ------- | ------- |
| [GRPC protocol][GRPC] | 5 | GRPC adds a 5-byte header for _Length-Prefixed-Messages_ |
| [PowerDNS logging][PDNS] | 2 | Sent messages contain a 2-byte header containing the message length |

[GRPC]: https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
[PDNS]: https://docs.powerdns.com/recursor/lua-config/protobuf.html

### Concise Binary Object Representation notes

Concise Binary Object Representation support numeric keys in the data. However,
XML (and this parser) expects node names to be strings starting with a letter.
To be compatible with these requirements, numeric nodes will be prefixed with
a lower case `n` and converted to strings. This means that if you for example
have a node with the key `123` in CBOR you will need to query `n123` in your
XPath expressions.

## Configuration

```toml
[[inputs.file]]
  files = ["example.xml"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "xml"

  ## PROTOCOL-BUFFER definitions
  ## Protocol-buffer definition file
  # xpath_protobuf_file = "sparkplug_b.proto"
  ## Name of the protocol-buffer message type to use in a fully qualified form.
  # xpath_protobuf_type = "org.eclipse.tahu.protobuf.Payload"
  ## List of paths to use when looking up imported protocol-buffer definition files.
  # xpath_protobuf_import_paths = ["."]
  ## Number of (header) bytes to ignore before parsing the message.
  # xpath_protobuf_skip_bytes = 0

  ## Print the internal XML document when in debug logging mode.
  ## This is especially useful when using the parser with non-XML formats like protocol-buffers
  ## to get an idea on the expression necessary to derive fields etc.
  # xpath_print_document = false

  ## Allow the results of one of the parsing sections to be empty.
  ## Useful when not all selected files have the exact same structure.
  # xpath_allow_empty_selection = false

  ## Get native data-types for all data-format that contain type information.
  ## Currently, CBOR, protobuf, msgpack and JSON support native data-types.
  # xpath_native_types = false

  ## Multiple parsing sections are allowed
  [[inputs.file.xpath]]
    ## Optional: XPath-query to select a subset of nodes from the XML document.
    # metric_selection = "/Bus/child::Sensor"

    ## Optional: XPath-query to set the metric (measurement) name.
    # metric_name = "string('example')"

    ## Optional: Query to extract metric timestamp.
    ## If not specified the time of execution is used.
    # timestamp = "/Gateway/Timestamp"
    ## Optional: Format of the timestamp determined by the query above.
    ## This can be any of "unix", "unix_ms", "unix_us", "unix_ns" or a valid Golang
    ## time format. If not specified, a "unix" timestamp (in seconds) is expected.
    # timestamp_format = "2006-01-02T15:04:05Z"
    ## Optional: Timezone of the parsed time
    ## This will locate the parsed time to the given timezone. Please note that
    ## for times with timezone-offsets (e.g. RFC3339) the timestamp is unchanged.
    ## This is ignored for all (unix) timestamp formats.
    # timezone = "UTC"

    ## Optional: List of fields to convert to hex-strings if they are
    ## containing byte-arrays. This might be the case for e.g. protocol-buffer
    ## messages encoding data as byte-arrays. Wildcard patterns are allowed.
    ## By default, all byte-array-fields are converted to string.
    # fields_bytes_as_hex = []

    ## Tag definitions using the given XPath queries.
    [inputs.file.xpath.tags]
      name   = "substring-after(Sensor/@name, ' ')"
      device = "string('the ultimate sensor')"

    ## Integer field definitions using XPath queries.
    [inputs.file.xpath.fields_int]
      consumers = "Variable/@consumers"

    ## Non-integer field definitions using XPath queries.
    ## The field type is defined using XPath expressions such as number(), boolean() or string(). If no conversion is performed the field will be of type string.
    [inputs.file.xpath.fields]
      temperature = "number(Variable/@temperature)"
      power       = "number(Variable/@power)"
      frequency   = "number(Variable/@frequency)"
      ok          = "Mode != 'ok'"
```

In this configuration mode, you explicitly specify the field and tags you want
to scrape out of your data.

A configuration can contain muliple _xpath_ subsections for e.g. the file plugin
to process the xml-string multiple times. Consult the [XPath syntax][xpath] and
the [underlying library's functions][xpath lib] for details and help regarding
XPath queries. Consider using an XPath tester such as [xpather.com][xpather] or
[Code Beautify's XPath Tester][xpath tester] for help developing and debugging
your query.

## Configuration (batch)

Alternatively to the configuration above, fields can also be specified in a
batch way. So contrary to specify the fields in a section, you can define a
`name` and a `value` selector used to determine the name and value of the fields
in the metric.

```toml
[[inputs.file]]
  files = ["example.xml"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "xml"

  ## PROTOCOL-BUFFER definitions
  ## Protocol-buffer definition file
  # xpath_protobuf_file = "sparkplug_b.proto"
  ## Name of the protocol-buffer message type to use in a fully qualified form.
  # xpath_protobuf_type = "org.eclipse.tahu.protobuf.Payload"
  ## List of paths to use when looking up imported protocol-buffer definition files.
  # xpath_protobuf_import_paths = ["."]

  ## Print the internal XML document when in debug logging mode.
  ## This is especially useful when using the parser with non-XML formats like protocol-buffers
  ## to get an idea on the expression necessary to derive fields etc.
  # xpath_print_document = false

  ## Allow the results of one of the parsing sections to be empty.
  ## Useful when not all selected files have the exact same structure.
  # xpath_allow_empty_selection = false

  ## Get native data-types for all data-format that contain type information.
  ## Currently, protobuf, msgpack and JSON support native data-types
  # xpath_native_types = false

  ## Multiple parsing sections are allowed
  [[inputs.file.xpath]]
    ## Optional: XPath-query to select a subset of nodes from the XML document.
    metric_selection = "/Bus/child::Sensor"

    ## Optional: XPath-query to set the metric (measurement) name.
    # metric_name = "string('example')"

    ## Optional: Query to extract metric timestamp.
    ## If not specified the time of execution is used.
    # timestamp = "/Gateway/Timestamp"
    ## Optional: Format of the timestamp determined by the query above.
    ## This can be any of "unix", "unix_ms", "unix_us", "unix_ns" or a valid Golang
    ## time format. If not specified, a "unix" timestamp (in seconds) is expected.
    # timestamp_format = "2006-01-02T15:04:05Z"

    ## Field specifications using a selector.
    field_selection = "child::*"
    ## Optional: Queries to specify field name and value.
    ## These options are only to be used in combination with 'field_selection'!
    ## By default the node name and node content is used if a field-selection
    ## is specified.
    # field_name  = "name()"
    # field_value = "."

    ## Optional: Expand field names relative to the selected node
    ## This allows to flatten out nodes with non-unique names in the subtree
    # field_name_expansion = false

    ## Tag specifications using a selector.
    ## tag_selection = "child::*"
    ## Optional: Queries to specify tag name and value.
    ## These options are only to be used in combination with 'tag_selection'!
    ## By default the node name and node content is used if a tag-selection
    ## is specified.
    # tag_name  = "name()"
    # tag_value = "."

    ## Optional: Expand tag names relative to the selected node
    ## This allows to flatten out nodes with non-unique names in the subtree
    # tag_name_expansion = false

    ## Tag definitions using the given XPath queries.
    [inputs.file.xpath.tags]
      name   = "substring-after(Sensor/@name, ' ')"
      device = "string('the ultimate sensor')"

```

__Please note__: The resulting fields are _always_ of type string!

It is also possible to specify a mixture of the two alternative ways of
specifying fields. In this case _explicitly_ defined tags and fields take
_precedence_ over the batch instances if both use the same tag/field name.

### metric_selection (optional)

You can specify a [XPath][xpath] query to select a subset of nodes from the XML
document, each used to generate a new metrics with the specified fields, tags
etc.

For relative queries in subsequent queries they are relative to the
`metric_selection`. To specify absolute paths, please start the query with a
slash (`/`).

Specifying `metric_selection` is optional. If not specified all relative queries
are relative to the root node of the XML document.

### metric_name (optional)

By specifying `metric_name` you can override the metric/measurement name with
the result of the given [XPath][xpath] query. If not specified, the default
metric name is used.

### timestamp, timestamp_format, timezone (optional)

By default the current time will be used for all created metrics. To set the
time from values in the XML document you can specify a [XPath][xpath] query in
`timestamp` and set the format in `timestamp_format`.

The `timestamp_format` can be set to `unix`, `unix_ms`, `unix_us`, `unix_ns`, or
an accepted [Go "reference time"][time const]. Consult the Go [time][time parse]
package for details and additional examples on how to set the time format.  If
`timestamp_format` is omitted `unix` format is assumed as result of the
`timestamp` query.

The `timezone` setting will be used to locate the parsed time in the given
timezone. This is helpful for cases where the time does not contain timezone
information, e.g. `2023-03-09 14:04:40` and is not located in _UTC_, which is
the default setting. It is also possible to set the `timezone` to `Local` which
used the configured host timezone.

For time formats with timezone information, e.g. RFC3339, the resulting
timestamp is unchanged. The `timezone` setting is ignored for all `unix`
timestamp formats.

### tags sub-section

[XPath][xpath] queries in the `tag name = query` format to add tags to the
metrics. The specified path can be absolute (starting with `/`) or
relative. Relative paths use the currently selected node as reference.

__NOTE:__ Results of tag-queries will always be converted to strings.

### fields_int sub-section

[XPath][xpath] queries in the `field name = query` format to add integer typed
fields to the metrics. The specified path can be absolute (starting with `/`) or
relative. Relative paths use the currently selected node as reference.

__NOTE:__ Results of field_int-queries will always be converted to
__int64__. The conversion will fail in case the query result is not convertible!

### fields sub-section

[XPath][xpath] queries in the `field name = query` format to add non-integer
fields to the metrics. The specified path can be absolute (starting with `/`) or
relative. Relative paths use the currently selected node as reference.

The type of the field is specified in the [XPath][xpath] query using the type
conversion functions of XPath such as `number()`, `boolean()` or `string()` If
no conversion is performed in the query the field will be of type string.

__NOTE: Path conversion functions will always succeed even if you convert a text
to float!__

### field_selection, field_name, field_value (optional)

You can specify a [XPath][xpath] query to select a set of nodes forming the
fields of the metric. The specified path can be absolute (starting with `/`) or
relative to the currently selected node. Each node selected by `field_selection`
forms a new field within the metric.

The _name_ and the _value_ of each field can be specified using the optional
`field_name` and `field_value` queries. The queries are relative to the selected
field if not starting with `/`. If not specified the field's _name_ defaults to
the node name and the field's _value_ defaults to the content of the selected
field node.

__NOTE__: `field_name` and `field_value` queries are only evaluated if a
`field_selection` is specified.

Specifying `field_selection` is optional. This is an alternative way to specify
fields especially for documents where the node names are not known a priori or
if there is a large number of fields to be specified. These options can also be
combined with the field specifications above.

__NOTE: Path conversion functions will always succeed even if you convert a text
to float!__

### field_name_expansion (optional)

When _true_, field names selected with `field_selection` are expanded to a
_path_ relative to the _selected node_. This is necessary if we e.g. select all
leaf nodes as fields and those leaf nodes do not have unique names. That is in
case you have duplicate names in the fields you select you should set this to
`true`.

### tag_selection, tag_name, tag_value (optional)

You can specify a [XPath][xpath] query to select a set of nodes forming the tags
of the metric. The specified path can be absolute (starting with `/`) or
relative to the currently selected node. Each node selected by `tag_selection`
forms a new tag within the metric.

The _name_ and the _value_ of each tag can be specified using the optional
`tag_name` and `tag_value` queries. The queries are relative to the selected tag
if not starting with `/`. If not specified the tag's _name_ defaults to the node
name and the tag's _value_ defaults to the content of the selected tag node.
__NOTE__: `tag_name` and `tag_value` queries are only evaluated if a
`tag_selection` is specified.

Specifying `tag_selection` is optional. This is an alternative way to specify
tags especially for documents where the node names are not known a priori or if
there is a large number of tags to be specified. These options can also be
combined with the tag specifications above.

### tag_name_expansion (optional)

When _true_, tag names selected with `tag_selection` are expanded to a _path_
relative to the _selected node_. This is necessary if we e.g. select all leaf
nodes as tags and those leaf nodes do not have unique names. That is in case you
have duplicate names in the tags you select you should set this to `true`.

## Examples

This `example.xml` file is used in the configuration examples below:

```xml
<?xml version="1.0"?>
<Gateway>
  <Name>Main Gateway</Name>
  <Timestamp>2020-08-01T15:04:03Z</Timestamp>
  <Sequence>12</Sequence>
  <Status>ok</Status>
</Gateway>

<Bus>
  <Sensor name="Sensor Facility A">
    <Variable temperature="20.0"/>
    <Variable power="123.4"/>
    <Variable frequency="49.78"/>
    <Variable consumers="3"/>
    <Mode>busy</Mode>
  </Sensor>
  <Sensor name="Sensor Facility B">
    <Variable temperature="23.1"/>
    <Variable power="14.3"/>
    <Variable frequency="49.78"/>
    <Variable consumers="1"/>
    <Mode>standby</Mode>
  </Sensor>
  <Sensor name="Sensor Facility C">
    <Variable temperature="19.7"/>
    <Variable power="0.02"/>
    <Variable frequency="49.78"/>
    <Variable consumers="0"/>
    <Mode>error</Mode>
  </Sensor>
</Bus>
```

### Basic Parsing

This example shows the basic usage of the xml parser.

Config:

```toml
[[inputs.file]]
  files = ["example.xml"]
  data_format = "xml"

  [[inputs.file.xpath]]
    [inputs.file.xpath.tags]
      gateway = "substring-before(/Gateway/Name, ' ')"

    [inputs.file.xpath.fields_int]
      seqnr = "/Gateway/Sequence"

    [inputs.file.xpath.fields]
      ok = "/Gateway/Status = 'ok'"
```

Output:

```text
file,gateway=Main,host=Hugin seqnr=12i,ok=true 1598610830000000000
```

In the _tags_ definition the XPath function `substring-before()` is used to only
extract the sub-string before the space. To get the integer value of
`/Gateway/Sequence` we have to use the _fields_int_ section as there is no XPath
expression to convert node values to integers (only float).

The `ok` field is filled with a boolean by specifying a query comparing the
query result of `/Gateway/Status` with the string _ok_. Use the type conversions
available in the XPath syntax to specify field types.

### Time and metric names

This is an example for using time and name of the metric from the XML document
itself.

Config:

```toml
[[inputs.file]]
  files = ["example.xml"]
  data_format = "xml"

  [[inputs.file.xpath]]
    metric_name = "name(/Gateway/Status)"

    timestamp = "/Gateway/Timestamp"
    timestamp_format = "2006-01-02T15:04:05Z"

    [inputs.file.xpath.tags]
      gateway = "substring-before(/Gateway/Name, ' ')"

    [inputs.file.xpath.fields]
      ok = "/Gateway/Status = 'ok'"
```

Output:

```text
Status,gateway=Main,host=Hugin ok=true 1596294243000000000
```

Additionally to the basic parsing example, the metric name is defined as the
name of the `/Gateway/Status` node and the timestamp is derived from the XML
document instead of using the execution time.

### Multi-node selection

For XML documents containing metrics for e.g. multiple devices (like `Sensor`s
in the _example.xml_), multiple metrics can be generated using node
selection. This example shows how to generate a metric for each _Sensor_ in the
example.

Config:

```toml
[[inputs.file]]
  files = ["example.xml"]
  data_format = "xml"

  [[inputs.file.xpath]]
    metric_selection = "/Bus/child::Sensor"

    metric_name = "string('sensors')"

    timestamp = "/Gateway/Timestamp"
    timestamp_format = "2006-01-02T15:04:05Z"

    [inputs.file.xpath.tags]
      name = "substring-after(@name, ' ')"

    [inputs.file.xpath.fields_int]
      consumers = "Variable/@consumers"

    [inputs.file.xpath.fields]
      temperature = "number(Variable/@temperature)"
      power       = "number(Variable/@power)"
      frequency   = "number(Variable/@frequency)"
      ok          = "Mode != 'error'"

```

Output:

```text
sensors,host=Hugin,name=Facility\ A consumers=3i,frequency=49.78,ok=true,power=123.4,temperature=20 1596294243000000000
sensors,host=Hugin,name=Facility\ B consumers=1i,frequency=49.78,ok=true,power=14.3,temperature=23.1 1596294243000000000
sensors,host=Hugin,name=Facility\ C consumers=0i,frequency=49.78,ok=false,power=0.02,temperature=19.7 1596294243000000000
```

Using the `metric_selection` option we select all `Sensor` nodes in the XML
document. Please note that all field and tag definitions are relative to these
selected nodes. An exception is the timestamp definition which is relative to
the root node of the XML document.

### Batch field processing with multi-node selection

For XML documents containing metrics with a large number of fields or where the
fields are not known before (e.g. an unknown set of `Variable` nodes in the
_example.xml_), field selectors can be used. This example shows how to generate
a metric for each _Sensor_ in the example with fields derived from the
_Variable_ nodes.

Config:

```toml
[[inputs.file]]
  files = ["example.xml"]
  data_format = "xml"

  [[inputs.file.xpath]]
    metric_selection = "/Bus/child::Sensor"
    metric_name = "string('sensors')"

    timestamp = "/Gateway/Timestamp"
    timestamp_format = "2006-01-02T15:04:05Z"

    field_selection = "child::Variable"
    field_name = "name(@*[1])"
    field_value = "number(@*[1])"

    [inputs.file.xpath.tags]
      name = "substring-after(@name, ' ')"
```

Output:

```text
sensors,host=Hugin,name=Facility\ A consumers=3,frequency=49.78,power=123.4,temperature=20 1596294243000000000
sensors,host=Hugin,name=Facility\ B consumers=1,frequency=49.78,power=14.3,temperature=23.1 1596294243000000000
sensors,host=Hugin,name=Facility\ C consumers=0,frequency=49.78,power=0.02,temperature=19.7 1596294243000000000
```

Using the `metric_selection` option we select all `Sensor` nodes in the XML
document. For each _Sensor_ we then use `field_selection` to select all child
nodes of the sensor as _field-nodes_ Please note that the field selection is
relative to the selected nodes.  For each selected _field-node_ we use
`field_name` and `field_value` to determining the field's name and value,
respectively. The `field_name` derives the name of the first attribute of the
node, while `field_value` derives the value of the first attribute and converts
the result to a number.

[cbor]:         https://cbor.io/
[json]:         https://www.json.org/
[msgpack]:      https://msgpack.org/
[protobuf]:     https://developers.google.com/protocol-buffers
[time const]:   https://golang.org/pkg/time/#pkg-constants
[time parse]:   https://golang.org/pkg/time/#Parse
[xml]:          https://www.w3.org/XML/
[xpath]:        https://www.w3.org/TR/xpath/
[xpath lib]:    https://github.com/antchfx/xpath
[xpath tester]: https://codebeautify.org/Xpath-Tester
[xpather]:      http://xpather.com/
