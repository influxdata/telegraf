# XML

The `xml` data format parses data in XML format.
This plugin using [etree package](https://github.com/beevik/etree)

### Configuration

```toml
[[inputs.file]]
  files = [ "data.xml" ]
  data_format = "xml"
  ##  xml_query parameter using XPath query for limits the list of 
  ##  analyzed nodes. Default - "//"
  xml_query = "//Node/"
  ##  xml_merge_nodes determines whether all extracted keys will be 
  ## merged into one metric. Default - false
  xml_merge_nodes = true
  ##  if xml_node_to_tag equals "true", the name of the node is recorded 
  ##  in the tags with the key "node_name". Default - false
  xml_node_to_tag = true
  ## Tag keys is an array of keys that should be added as tags.
  ## Matching keys are no longer saved as fields.
  tag_keys = [
    "my_tag_1",
    "my_tag_2"
  ]
```

#### xml_merge_nodes

This determines whether all extracted keys will be merged into one metric.  
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
For a previous input example, if **xml_node_to_tag=true**:
```
file,node_name=Host_1 Name="Host",ConnectionsCurrent=0i,ConnectionsTotal=0i
file,node_name=Host_2 Name="Server",ConnectionsCurrent=0i,ConnectionsTotal=0i
```

### Metrics

If the node or attribute value contains only *\s*, *\t*, *\r* or *\n* characters, it will be discarded.  
Each value sequentially passes conversion attempts to Int64, Float64 and Boolean using **strconv**.  
If the conversion was successful, the result of the conversion is written in the field, otherwise a string is returned.
