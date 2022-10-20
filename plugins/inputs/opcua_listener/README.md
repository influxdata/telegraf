# OPC UA Client Listener Input Plugin

The `opcua_listener` plugin subscribes to data from OPC UA Server devices.

Telegraf minimum version: Telegraf 1.25
Plugin minimum tested version: 1.25

## Configuration

```toml @sample.conf
# Retrieve data from OPCUA devices
[[inputs.opcua_listener]]
  ## Metric name
  # name = "opcua_listener"
  #
  ## OPC UA Endpoint URL
  # endpoint = "opc.tcp://localhost:4840"
  #
  ## Maximum time allowed to establish a connect to the endpoint.
  # connect_timeout = "10s"
  #
  ## Maximum time allowed for a request over the established connection.
  # request_timeout = "5s"
  #
  ## The interval at which the server should at least update its monitored items
  # subscription_interval = "100ms"
  #
  ## Security policy, one of "None", "Basic128Rsa15", "Basic256",
  ## "Basic256Sha256", or "auto"
  # security_policy = "auto"
  #
  ## Security mode, one of "None", "Sign", "SignAndEncrypt", or "auto"
  # security_mode = "auto"
  #
  ## Path to cert.pem. Required when security mode or policy isn't "None".
  ## If cert path is not supplied, self-signed cert and key will be generated.
  # certificate = "/etc/telegraf/cert.pem"
  #
  ## Path to private key.pem. Required when security mode or policy isn't "None".
  ## If key path is not supplied, self-signed cert and key will be generated.
  # private_key = "/etc/telegraf/key.pem"
  #
  ## Authentication Method, one of "Certificate", "UserName", or "Anonymous".  To
  ## authenticate using a specific ID, select 'Certificate' or 'UserName'
  # auth_method = "Anonymous"
  #
  ## Username. Required for auth_method = "UserName"
  # username = ""
  #
  ## Password. Required for auth_method = "UserName"
  # password = ""
  #
  ## Option to select the metric timestamp to use. Valid options are:
  ##     "gather" -- uses the time of receiving the data in telegraf
  ##     "server" -- uses the timestamp provided by the server
  ##     "source" -- uses the timestamp provided by the source
  # timestamp = "gather"
  #
  ## Node ID configuration
  ## name              - field name to use in the output
  ## namespace         - OPC UA namespace of the node (integer value 0 thru 3)
  ## identifier_type   - OPC UA ID type (s=string, i=numeric, g=guid, b=opaque)
  ## identifier        - OPC UA ID (tag as shown in opcua browser)
  ## default_tags      - extra tags to be added to the output metric (optional)
  ##
  ## Use either the inline notation or the bracketed notation, not both.
  #
  ## Inline notation (default_tags not supported yet)
  # nodes = [
  #   {name="", namespace="", identifier_type="", identifier=""},
  #   {name="", namespace="", identifier_type="", identifier=""},
  # ]
  #
  ## Bracketed notation
  # [[inputs.opcua_listener.nodes]]
  #   name = "node1"
  #   namespace = ""
  #   identifier_type = ""
  #   identifier = ""
  #   default_tags = { tag1 = "value1", tag2 = "value2" }
  #
  # [[inputs.opcua_listener.nodes]]
  #   name = "node2"
  #   namespace = ""
  #   identifier_type = ""
  #   identifier = ""
  #
  ## Node Group
  ## Sets defaults so they aren't required in every node.
  ## Default values can be set for:
  ## * Metric name
  ## * OPC UA namespace
  ## * Identifier
  ## * Default tags
  ##
  ## Multiple node groups are allowed
  #[[inputs.opcua_listener.group]]
  ## Group Metric name. Overrides the top level name.  If unset, the
  ## top level name is used.
  # name =
  #
  ## Group default namespace. If a node in the group doesn't set its
  ## namespace, this is used.
  # namespace =
  #
  ## Group default identifier type. If a node in the group doesn't set its
  ## namespace, this is used.
  # identifier_type =
  #
  ## Default tags that are applied to every node in this group. Can be
  ## overwritten in a node by setting a different value for the tag name.
  ##   example: default_tags = { tag1 = "value1" }
  # default_tags = {}
  #
  ## Node ID Configuration.  Array of nodes with the same settings as above.
  ## Use either the inline notation or the bracketed notation, not both.
  #
  ## Inline notation (default_tags not supported yet)
  # nodes = [
  #  {name="node1", namespace="", identifier_type="", identifier=""},
  #  {name="node2", namespace="", identifier_type="", identifier=""},
  #]
  #
  ## Bracketed notation
  # [[inputs.opcua_listener.group.nodes]]
  #   name = "node1"
  #   namespace = ""
  #   identifier_type = ""
  #   identifier = ""
  #   default_tags = { tag1 = "override1", tag2 = "value2" }
  #
  # [[inputs.opcua_listener.group.nodes]]
  #   name = "node2"
  #   namespace = ""
  #   identifier_type = ""
  #   identifier = ""

  ## Enable workarounds required by some devices to work correctly
  # [inputs.opcua_listener.workarounds]
    ## Set additional valid status codes, StatusOK (0x0) is always considered valid
    # additional_valid_status_codes = ["0xC0"]

  # [inputs.opcua_listener.request_workarounds]
    ## Use unregistered reads instead of registered reads
    # use_unregistered_reads = false
```

## Node Configuration

An OPC UA node ID may resemble: "ns=3;s=Temperature". In this example:

- ns=3 is indicating the `namespace` is 3
- s=Temperature is indicting that the `identifier_type` is a string and `identifier` value is 'Temperature'
- This example temperature node has a value of 79.0
To gather data from this node enter the following line into the 'nodes' property above:

```text
{field_name="temp", namespace="3", identifier_type="s", identifier="Temperature"},
```

This node configuration produces a metric like this:

```text
opcua,id=ns\=3;s\=Temperature temp=79.0,quality="OK (0x0)" 1597820490000000000
```

## Group Configuration

Groups can set default values for the namespace, identifier type, and
tags settings.  The default values apply to all the nodes in the
group.  If a default is set, a node may omit the setting altogether.
This simplifies node configuration, especially when many nodes share
the same namespace or identifier type.

The output metric will include tags set in the group and the node.  If
a tag with the same name is set in both places, the tag value from the
node is used.

This example group configuration has three groups with two nodes each:

```toml
  # Group 1
  [[inputs.opcua_listener.group]]
    name = "group1_metric_name"
    namespace = "3"
    identifier_type = "i"
    default_tags = { group1_tag = "val1" }
    [[inputs.opcua.group.nodes]]
      name = "name"
      identifier = "1001"
      default_tags = { node1_tag = "val2" }
    [[inputs.opcua.group.nodes]]
      name = "name"
      identifier = "1002"
      default_tags = {node1_tag = "val3"}
  
  # Group 2
  [[inputs.opcua_listener.group]]
    name = "group2_metric_name"
    namespace = "3"
    identifier_type = "i"
    default_tags = { group2_tag = "val3" }
    [[inputs.opcua.group.nodes]]
      name = "saw"
      identifier = "1003"
      default_tags = { node2_tag = "val4" }
    [[inputs.opcua.group.nodes]]
      name = "sin"
      identifier = "1004"
  
  # Group 3
  [[inputs.opcua_listener.group]]
    name = "group3_metric_name"
    namespace = "3"
    identifier_type = "i"
    default_tags = { group3_tag = "val5" }
    nodes = [
      {name="name", identifier="1001"},
      {name="name", identifier="1002"},
    ]
```

## Connection Service

This plugin subscribes to the specified nodes to receive data from
the OPC server. The updates are received at most as fast as the
`subscription_interval`.

## Metrics

The metrics collected by this input plugin will depend on the
configured `nodes` and `group`.

## Example Output

```text
group1_metric_name,group1_tag=val1,id=ns\=3;i\=1001,node1_tag=val2 name=0,Quality="OK (0x0)" 1606893246000000000
group1_metric_name,group1_tag=val1,id=ns\=3;i\=1002,node1_tag=val3 name=-1.389117,Quality="OK (0x0)" 1606893246000000000
group2_metric_name,group2_tag=val3,id=ns\=3;i\=1003,node2_tag=val4 Quality="OK (0x0)",saw=-1.6 1606893246000000000
group2_metric_name,group2_tag=val3,id=ns\=3;i\=1004 sin=1.902113,Quality="OK (0x0)" 1606893246000000000
```
