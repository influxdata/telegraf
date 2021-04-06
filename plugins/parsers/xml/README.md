# XML

The XML data format parser parses a [XML][xml] string into metric fields using [XPath][xpath] expressions. For supported
XPath functions check [the underlying XPath library][xpath lib].

**NOTE:** The type of fields are specified using [XPath functions][xpath lib]. The only exception are *integer* fields
that need to be specified in a `fields_int` section.

### Configuration

```toml
[[inputs.file]]
  files = ["example.xml"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "xml"

  ## Multiple parsing sections are allowed
  [[inputs.file.xml]]
    ## Optional: XPath-query to select a subset of nodes from the XML document.
    #metric_selection = "/Bus/child::Sensor"

    ## Optional: XPath-query to set the metric (measurement) name.
    #metric_name = "string('example')"

    ## Optional: Query to extract metric timestamp.
    ## If not specified the time of execution is used.
    #timestamp = "/Gateway/Timestamp"
    ## Optional: Format of the timestamp determined by the query above.
    ## This can be any of "unix", "unix_ms", "unix_us", "unix_ns" or a valid Golang
    ## time format. If not specified, a "unix" timestamp (in seconds) is expected.
    #timestamp_format = "2006-01-02T15:04:05Z"

    ## Tag definitions using the given XPath queries.
    [inputs.file.xml.tags]
      name   = "substring-after(Sensor/@name, ' ')"
      device = "string('the ultimate sensor')"

    ## Integer field definitions using XPath queries.
    [inputs.file.xml.fields_int]
      consumers = "Variable/@consumers"

    ## Non-integer field definitions using XPath queries.
    ## The field type is defined using XPath expressions such as number(), boolean() or string(). If no conversion is performed the field will be of type string.
    [inputs.file.xml.fields]
      temperature = "number(Variable/@temperature)"
      power       = "number(Variable/@power)"
      frequency   = "number(Variable/@frequency)"
      ok          = "Mode != 'ok'"
```

A configuration can contain muliple *xml* subsections for e.g. the file plugin to process the xml-string multiple times.
Consult the [XPath syntax][xpath] and the [underlying library's functions][xpath lib] for details and help regarding XPath queries. Consider using an XPath tester such as [xpather.com][xpather] or [Code Beautify's XPath Tester][xpath tester] for help developing and debugging 
your query.

Alternatively to the configuration above, fields can also be specified in a batch way. So contrary to specify the fields
in a section, you can define a `name` and a `value` selector used to determine the name and value of the fields in the
metric.
```toml
[[inputs.file]]
  files = ["example.xml"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "xml"

  ## Multiple parsing sections are allowed
  [[inputs.file.xml]]
    ## Optional: XPath-query to select a subset of nodes from the XML document.
    metric_selection = "/Bus/child::Sensor"

    ## Optional: XPath-query to set the metric (measurement) name.
    #metric_name = "string('example')"

    ## Optional: Query to extract metric timestamp.
    ## If not specified the time of execution is used.
    #timestamp = "/Gateway/Timestamp"
    ## Optional: Format of the timestamp determined by the query above.
    ## This can be any of "unix", "unix_ms", "unix_us", "unix_ns" or a valid Golang
    ## time format. If not specified, a "unix" timestamp (in seconds) is expected.
    #timestamp_format = "2006-01-02T15:04:05Z"

    ## Field specifications using a selector.
    field_selection = "child::*"
    ## Optional: Queries to specify field name and value.
    ## These options are only to be used in combination with 'field_selection'!
    ## By default the node name and node content is used if a field-selection
    ## is specified.
    #field_name  = "name()"
    #field_value = "."

    ## Optional: Expand field names relative to the selected node
    ## This allows to flatten out nodes with non-unique names in the subtree
    #field_name_expansion = false

    ## Tag definitions using the given XPath queries.
    [inputs.file.xml.tags]
      name   = "substring-after(Sensor/@name, ' ')"
      device = "string('the ultimate sensor')"

```
*Please note*: The resulting fields are _always_ of type string!

It is also possible to specify a mixture of the two alternative ways of specifying fields.

#### metric_selection (optional)

You can specify a [XPath][xpath] query to select a subset of nodes from the XML document, each used to generate a new
metrics with the specified fields, tags etc.

For relative queries in subsequent queries they are relative to the `metric_selection`. To specify absolute paths, please start the query with a slash (`/`).

Specifying `metric_selection` is optional. If not specified all relative queries are relative to the root node of the XML document.

#### metric_name (optional)

By specifying `metric_name` you can override the metric/measurement name with the result of the given [XPath][xpath] query. If not specified, the default metric name is used.

#### timestamp, timestamp_format (optional)

By default the current time will be used for all created metrics. To set the time from values in the XML document you can specify a [XPath][xpath] query in `timestamp` and set the format in `timestamp_format`.

The `timestamp_format` can be set to `unix`, `unix_ms`, `unix_us`, `unix_ns`, or
an accepted [Go "reference time"][time const]. Consult the Go [time][time parse] package for details and additional examples on how to set the time format.
If `timestamp_format` is omitted `unix` format is assumed as result of the `timestamp` query.

#### tags sub-section

[XPath][xpath] queries in the `tag name = query` format to add tags to the metrics. The specified path can be absolute (starting with `/`) or relative. Relative paths use the currently selected node as reference.

**NOTE:** Results of tag-queries will always be converted to strings.

#### fields_int sub-section

[XPath][xpath] queries in the `field name = query` format to add integer typed fields to the metrics. The specified path can be absolute (starting with `/`) or relative. Relative paths use the currently selected node as reference.

**NOTE:** Results of field_int-queries will always be converted to **int64**. The conversion will fail in case the query result is not convertible!

#### fields sub-section

[XPath][xpath] queries in the `field name = query` format to add non-integer fields to the metrics. The specified path can be absolute (starting with `/`) or relative. Relative paths use the currently selected node as reference.

The type of the field is specified in the [XPath][xpath] query using the type conversion functions of XPath such as `number()`, `boolean()` or `string()`
If no conversion is performed in the query the field will be of type string.

**NOTE: Path conversion functions will always succeed even if you convert a text to float!**


#### field_selection, field_name, field_value (optional)

You can specify a [XPath][xpath] query to select a set of nodes forming the fields of the metric. The specified path can be absolute (starting with `/`) or relative to the currently selected node. Each node selected by `field_selection` forms a new field within the metric.

The *name* and the *value* of each field can be specified using the optional `field_name` and `field_value` queries. The queries are relative to the selected field if not starting with `/`. If not specified the field's *name* defaults to the node name and the field's *value* defaults to the content of the selected field node.
**NOTE**: `field_name` and `field_value` queries are only evaluated if a `field_selection` is specified.

Specifying `field_selection` is optional. This is an alternative way to specify fields especially for documents where the node names are not known a priori or if there is a large number of fields to be specified. These options can also be combined with the field specifications above.

**NOTE: Path conversion functions will always succeed even if you convert a text to float!**

#### field_name_expansion (optional)

When *true*, field names selected with `field_selection` are expanded to a *path* relative to the *selected node*. This
is necessary if we e.g. select all leaf nodes as fields and those leaf nodes do not have unique names. That is in case
you have duplicate names in the fields you select you should set this to `true`.

### Examples

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

#### Basic Parsing

This example shows the basic usage of the xml parser.

Config:
```toml
[[inputs.file]]
  files = ["example.xml"]
  data_format = "xml"

  [[inputs.file.xml]]
    [inputs.file.xml.tags]
      gateway = "substring-before(/Gateway/Name, ' ')"

    [inputs.file.xml.fields_int]
      seqnr = "/Gateway/Sequence"

    [inputs.file.xml.fields]
      ok = "/Gateway/Status = 'ok'"
```

Output:
```
file,gateway=Main,host=Hugin seqnr=12i,ok=true 1598610830000000000
```

In the *tags* definition the XPath function `substring-before()` is used to only extract the sub-string before the space. To get the integer value of `/Gateway/Sequence` we have to use the *fields_int* section as there is no XPath expression to convert node values to integers (only float).
The `ok` field is filled with a boolean by specifying a query comparing the query result of `/Gateway/Status` with the string *ok*. Use the type conversions available in the XPath syntax to specify field types.

#### Time and metric names

This is an example for using time and name of the metric from the XML document itself.

Config:
```toml
[[inputs.file]]
  files = ["example.xml"]
  data_format = "xml"

  [[inputs.file.xml]]
    metric_name = "name(/Gateway/Status)"

    timestamp = "/Gateway/Timestamp"
    timestamp_format = "2006-01-02T15:04:05Z"

    [inputs.file.xml.tags]
      gateway = "substring-before(/Gateway/Name, ' ')"

    [inputs.file.xml.fields]
      ok = "/Gateway/Status = 'ok'"
```

Output:
```
Status,gateway=Main,host=Hugin ok=true 1596294243000000000
```
Additionally to the basic parsing example, the metric name is defined as the name of the `/Gateway/Status` node and the timestamp is derived from the XML document instead of using the execution time.

#### Multi-node selection

For XML documents containing metrics for e.g. multiple devices (like `Sensor`s in the *example.xml*), multiple metrics can be generated using node selection. This example shows how to generate a metric for each *Sensor* in the example.

Config:
```toml
[[inputs.file]]
  files = ["example.xml"]
  data_format = "xml"

  [[inputs.file.xml]]
    metric_selection = "/Bus/child::Sensor"

    metric_name = "string('sensors')"

    timestamp = "/Gateway/Timestamp"
    timestamp_format = "2006-01-02T15:04:05Z"

    [inputs.file.xml.tags]
      name = "substring-after(@name, ' ')"

    [inputs.file.xml.fields_int]
      consumers = "Variable/@consumers"

    [inputs.file.xml.fields]
      temperature = "number(Variable/@temperature)"
      power       = "number(Variable/@power)"
      frequency   = "number(Variable/@frequency)"
      ok          = "Mode != 'error'"

```

Output:
```
sensors,host=Hugin,name=Facility\ A consumers=3i,frequency=49.78,ok=true,power=123.4,temperature=20 1596294243000000000
sensors,host=Hugin,name=Facility\ B consumers=1i,frequency=49.78,ok=true,power=14.3,temperature=23.1 1596294243000000000
sensors,host=Hugin,name=Facility\ C consumers=0i,frequency=49.78,ok=false,power=0.02,temperature=19.7 1596294243000000000
```

Using the `metric_selection` option we select all `Sensor` nodes in the XML document. Please note that all field and tag definitions are relative to these selected nodes. An exception is the timestamp definition which is relative to the root node of the XML document.

#### Batch field processing with multi-node selection

For XML documents containing metrics with a large number of fields or where the fields are not known before (e.g. an unknown set of `Variable` nodes in the *example.xml*), field selectors can be used. This example shows how to generate a metric for each *Sensor* in the example with fields derived from the *Variable* nodes.

Config:
```toml
[[inputs.file]]
  files = ["example.xml"]
  data_format = "xml"

  [[inputs.file.xml]]
    metric_selection = "/Bus/child::Sensor"
    metric_name = "string('sensors')"

    timestamp = "/Gateway/Timestamp"
    timestamp_format = "2006-01-02T15:04:05Z"

    field_selection = "child::Variable"
    field_name = "name(@*[1])"
    field_value = "number(@*[1])"

    [inputs.file.xml.tags]
      name = "substring-after(@name, ' ')"
```

Output:
```
sensors,host=Hugin,name=Facility\ A consumers=3,frequency=49.78,power=123.4,temperature=20 1596294243000000000
sensors,host=Hugin,name=Facility\ B consumers=1,frequency=49.78,power=14.3,temperature=23.1 1596294243000000000
sensors,host=Hugin,name=Facility\ C consumers=0,frequency=49.78,power=0.02,temperature=19.7 1596294243000000000
```

Using the `metric_selection` option we select all `Sensor` nodes in the XML document. For each *Sensor* we then use `field_selection` to select all child nodes of the sensor as *field-nodes* Please note that the field selection is relative to the selected nodes.
For each selected *field-node* we use `field_name` and `field_value` to determining the field's name and value, respectively. The `field_name` derives the name of the first attribute of the node, while `field_value` derives the value of the first attribute  and converts the result to a number.

[xpath lib]:    https://github.com/antchfx/xpath
[xml]:          https://www.w3.org/XML/
[xpath]:        https://www.w3.org/TR/xpath/
[xpather]:      http://xpather.com/
[xpath tester]: https://codebeautify.org/Xpath-Tester
[time const]:   https://golang.org/pkg/time/#pkg-constants
[time parse]:   https://golang.org/pkg/time/#Parse
