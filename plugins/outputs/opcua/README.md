# OPC UA Output Plugin

This plugin writes to a (list of) specified OPC UA node(s).

```toml
[[outputs.opcua]]
  endpoint = "opc.tcp://localhost:49320"
  policy = "None"
  mode = "None"
  [outputs.opcua.node_id_map]
	usage_idle = "ns=2;s=MyPLC.NodeToUpdate"
	
  # node_id_map takes the form:
  #
  #[outputs.opcua.node_id_map]
  #  meric_field_name1 = "opcua id 1"
  #  meric_field_name2 = "opcua id 2"
  # 
  # The OPC UA client will iterate over the fields in a receieved metric and update the corresponding opcua node id with the metric field's value.
  #
  #
  # Full list of options:
  #
  #
  #	endpoint = "" #defaults to "opc.tcp://localhost:50000"
  # [node_id_map] #required
  #   field = "node id"
  #
  # policy = "" #defaults to "Auto"
  # mode = "" #defaults to "Auto"
  # username = "" #defaults to nil
  # password = "" #defaults to nil
  # cert_file = "" #defaults to "" - path to cert file
  # key_file = "" #defaults to "" - path to key file
  # auth_method = "" #defaults to "Anonymous" - accepts Anonymous, Username, Certificate
  ```