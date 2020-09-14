# XML

The `xml` data format parses data in XML format.
This plugin using [etree package](https://github.com/beevik/etree)

### Configuration

```toml
[[inputs.file]]
  files = [ "data.xml" ]
  data_format = "xml"
  
  ##  xml_query parameter using XPath-like query for limits the list of 
  ##  analyzed nodes. Default - "//"
  xml_query = "//Node/"

  ##  This parameter specifies which prefix will be used for the attribute 
  ##  when constructing the path to it. Default - "@"
  xml_attr_prefix = "_"

  ##  xml_merge_nodes determines whether all extracted keys will be 
  ##  merged into one metric. Default - false
  xml_merge_nodes = true
  
  ##  Determines whether the nodes should be parsed as array elements
  ##  When true, each node is analyzed separately and forms its own metric
  ##  This parameter changes the parser behavior. Default - false
  xml_array = true
  
  ##  If xml_node_to_tag equals "true", the name of the node is recorded 
  ##  in the tags with the key "xml_node_name". Default - false
  xml_node_to_tag = true

  ##  Indicates whether the parser will dynamically determine the data type 
  ##  in an element. Default - true
  xml_type_detection = true
    
  ##  Tag keys is an array of keys that should be added as tags.
  ##  Matching keys are no longer saved as fields.
  tag_keys = [
    "my_tag_1",
    "my_tag_2"
  ]
```

#### xml_merge_nodes

This parameter determines whether all extracted keys will be merged into one metric.  
For example, if your XML looks like this:
```xml
<VHost>
  <ConnectionsCurrent>0</ConnectionsCurrent>
  <ConnectionsTotal>0</ConnectionsTotal>
</VHost>
```
with configuration:
```toml
[[inputs.file]]
  files = [ "data.xml" ]
  data_format = "xml"
  xml_merge_nodes = true
  xml_query = "/VHost/*"

```
you will get this metric:
```
file ConnectionsCurrent=0i,ConnectionsTotal=0i
```
therwise, each node will be written in a separate row:
```
file ConnectionsCurrent=0i
file ConnectionsTotal=0i
```

Setting a parameter to false can be useful, if your data is combined in one node and declared in its attributes:
```xml
<VHost>
  <Host_1 Name="Host" ConnectionsCurrent="0" ConnectionsTotal="0" />
  <Host_2 Name="Server" ConnectionsCurrent="0" ConnectionsTotal="0" />
</VHost>
```
Output with **xml_merge_nodes = false**:
```
file Host_1@AvgCPU=13.3,Host_1@FQDN="host.local",Host_1@IsMaster=true,Host_1@Name="Host" 1600030888000000000
file Host_2@ConnectionsCurrent=5i,Host_2@ConnectionsTotal=18i,Host_2@Name="Server" 1600030888000000000
```

#### xml_node_to_tag
This parameter determines whether the node name will be added to the tags.  
For a previous input example, if **xml_node_to_tag = true**:
```
file,xml_node_name=Host_1 Host_1@AvgCPU=13.3,Host_1@FQDN="host.local",Host_1@IsMaster=true,Host_1@Name="Host" 1600030888000000000
file,xml_node_name=Host_2 Host_2@ConnectionsCurrent=5i,Host_2@ConnectionsTotal=18i,Host_2@Name="Server" 1600030888000000000
```

Note: to remove the prefix of the node name in the tags or field keys, you can use the [starlark processor](../../../plugins/processors/starlark).

#### xml_array
This parameter changes the parsing logic.  
When **xml_array = true**, parser will parse content and generate metric of each node separately.  
It\`s important that the XPath query must return an array of top-level nodes.  
  
The first situation is when your document looks like this:
```xml
<Document>
  <Data>
    <Host>
      <Name>Host_1</Name>
      <Uptime>1000</Uptime>
      <Connections>
        <Total>15</Total>
        <Current>2</Current>
      </Connections>
    </Host>
    <Host>
      <Name>Host_2</Name>
      <Uptime>1240</Uptime>
      <Connections>
        <Total>33</Total>
        <Current>4</Current>
      </Connections>
    </Host>
  </Data>
</Document>
```
With configuration:
```toml
[[inputs.file]]
  files = [ "data.xml" ]
  data_format = "xml"
  xml_query = "//Data/Host"
  xml_array = true
  tag_keys = [ "Name" ]
```
The result will be like this:
```
file,Name=Host_1 Connections_Current=2i,Connections_Total=15i,Uptime=1000i 1600031193000000000
file,Name=Host_2 Connections_Current=4i,Connections_Total=33i,Uptime=1240i 1600031193000000000
```
  
Another possible situation where the node names are different, but it is still an array:
```xml
<Document>
  <Data>
    <Host_1>
      <Name>Host_1</Name>
      <Uptime>1000</Uptime>
      <Connections>
        <Total>15</Total>
        <Current>2</Current>
      </Connections>
    </Host_1>
    <Host_2>
      <Name>Host_2</Name>
      <Uptime>1240</Uptime>
      <Connections>
        <Total>33</Total>
        <Current>4</Current>
      </Connections>
    </Host_2>
  </Data>
</Document>
```
For a similar result, a configuration like this will work for you:
```toml
[[inputs.file]]
  files = [ "data.xml" ]
  data_format = "xml"
  xml_query = "//Data/*"
  xml_node_to_tag = true
  xml_array = true
  tag_keys = [ "Name" ]
```
Result:
```
file,Name=Host_1,xml_node_name=Host_1 Connections_Current=2i,Connections_Total=15i,Uptime=1000i 1598638060000000000
file,Name=Host_2,xml_node_name=Host_2 Connections_Current=4i,Connections_Total=33i,Uptime=1240i 1598638060000000000
```

You can find more information about XPath queries in the [documentation for the etree package](https://pkg.go.dev/github.com/beevik/etree?tab=doc#Path).  

#### xml_type_detection
By default, each value sequentially passes conversion attempts to Int64, Float64 and Boolean using **strconv**. If the conversion was successful, the result of the conversion is written in the field, otherwise the string is returned.
  
Since XML itself does not provide the ability to unambiguously determine the data type, you may want to control this yourself.  
When this parameter is `false`, all fields are written as strings. After that, you can convert each field to the type you want using [converter processor](../../../plugins/processors/converter).


### Metrics
If the node or attribute value contains only *\s*, *\t*, *\r* or *\n* characters, it will be discarded.
  

When composing a tag or field name, the element's index in the tree is added to the path. So, if several nodes with the same name are found at the same level, an index will be added to the key. Index counting starts from zero.  
Example:
```xml
<?xml version="1.0"?>
<Gateway>
  <Name attr="NEW_NAME">Main_Gateway</Name>
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
Configuration:
```toml
[[inputs.file]]
  files = [ "data.xml" ]
  data_format = "xml"
  xml_query = "/Bus/Sensor"
  xml_array = true
  tag_keys = [ "@name" ]
```
Output:
```
file,@name=Sensor\ Facility\ A Mode=1i,Variable_0@temperature=20,Variable_1@power=123.4,Variable_2@frequency=49.78,Variable_3@consumers=3i 1600063993000000000
file,@name=Sensor\ Facility\ B Mode=1i,Variable_0@temperature=23.1,Variable_1@power=14.3,Variable_2@frequency=49.78,Variable_3@consumers=1i 1600063993000000000
file,@name=Sensor\ Facility\ C Mode=0i,Variable_0@temperature=19.7,Variable_1@power=0.02,Variable_2@frequency=49.78,Variable_3@consumers=0i 1600063993000000000
```
  
If your XML document is complex, you can try using [execd processor plugin](../../../plugins/processors/execd). Get your document through the [value parser](../value), and then use custom processor to retrieve data.

