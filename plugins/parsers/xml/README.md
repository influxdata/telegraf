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
  ##  xml_merge_nodes determines whether all extracted keys will be 
  ## merged into one metric. Default - false
  xml_merge_nodes = true
  ##  determines whether the nodes should be parsed as array elements
  ##  when true, each node is analyzed separately and forms its own metric
  ##  this parameter overrides the parser behavior. Default - false
  xml_array = true
  ##  if xml_node_to_tag equals "true", the name of the node is recorded 
  ##  in the tags with the key "xml_node_name". Default - false
  xml_node_to_tag = true
  ## Tag keys is an array of keys that should be added as tags.
  ## Matching keys are no longer saved as fields.
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
and **xml_merge_nodes = true**, you will get this metric:
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
file Name="Host",ConnectionsCurrent=0i,ConnectionsTotal=0i
file Name="Server",ConnectionsCurrent=0i,ConnectionsTotal=0i
```

#### xml_node_to_tag
This parameter determines whether the node name will be added to the tags.  
For a previous input example, if **xml_node_to_tag = true**:
```
file,xml_node_name=Host_1 Name="Host",ConnectionsCurrent=0i,ConnectionsTotal=0i
file,xml_node_name=Host_2 Name="Server",ConnectionsCurrent=0i,ConnectionsTotal=0i
```

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
file,Name=Host_1 Uptime=1000i,Total=15i,Current=2i 1598637420000000000
file,Name=Host_2 Uptime=1240i,Total=33i,Current=4i 1598637420000000000
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
file,Name=Host_1,xml_node_name=Host_1 Uptime=1000i,Total=15i,Current=2i 1598638060000000000
file,Name=Host_2,xml_node_name=Host_2 Uptime=1240i,Total=33i,Current=4i 1598638060000000000
```
  
You can find more information about XPath queries in the [documentation for the etree package](https://pkg.go.dev/github.com/beevik/etree?tab=doc#Path).  
  
If your XML document is complex, you can try using [execd processor plugin](../../../plugins/processors/execd). Get your document through the [value parser](../value), and then use custom processor to retrieve data.

### Metrics

If the node or attribute value contains only *\s*, *\t*, *\r* or *\n* characters, it will be discarded.  
Each value sequentially passes conversion attempts to Int64, Float64 and Boolean using **strconv**.  
If the conversion was successful, the result of the conversion is written in the field, otherwise the string is returned.
