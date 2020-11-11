# OPC UA Client Input Plugin

The `opcua` plugin retrieves data from OPC UA client devices.

Telegraf minimum version: Telegraf 1.16
Plugin minimum tested version: 1.16

### Configuration:

```toml
[[inputs.opcua]]
  ## Metric name
  # metric_name = "opcua"
  #
  ## OPC UA Endpoint URL
  # endpoint = "opc.tcp://localhost:4840"
  #
  ## Maximum time allowed to establish a connect to the endpoint.
  # connect_timeout = "10s"
  #
  ## Maximum time allowed for a request over the estabilished connection.
  # request_timeout = "5s"
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
  ## Node ID configuration
  ## field_name        - field name to use in the output
  ## namespace         - OPC UA namespace of the node (integer value 0 thru 3)
  ## identifier_type   - OPC UA ID type (s=string, i=numeric, g=guid, b=opaque)
  ## identifier        - OPC UA ID (tag as shown in opcua browser)
  ## Example:
  ## {name="ProductUri", namespace="0", identifier_type="i", identifier="2262"}
  # nodes = [
  #  {field_name="", namespace="", identifier_type="", identifier=""},
  #  {field_name="", namespace="", identifier_type="", identifier=""},
  #]
  #
  ## Node Group
  ## Sets defaults for OPC UA namespace and ID type so they aren't required in
  ## every node.  A group can also have a metric name that overrides the main
  ## plugin metric name.
  ##
  ## Multiple node groups are allowed
  #[[inputs.opcua.group]]
  ## Group Metric name. Overrides the top level metric_name.  If unset, the
  ## top level metric_name is used.
  # metric_name =
  #
  ## Group default namespace. If a node in the group doesn't set its
  ## namespace, this is used.
  # namespace =
  #
  ## Group default identifier type. If a node in the group doesn't set its
  ## namespace, this is used.
  # identifier_type =
  #
  ## Node ID Configuration.  Array of nodes with the same settings as above.
  # nodes = [
  #  {field_name="", namespace="", identifier_type="", identifier=""},
  #  {field_name="", namespace="", identifier_type="", identifier=""},
  #]
```

### Example Node Configuration
An OPC UA node ID may resemble: "n=3;s=Temperature". In this example:
- n=3 is indicating the `namespace` is 3
- s=Temperature is indicting that the `identifier_type` is a string and `identifier` value is 'Temperature'
- This example temperature node has a value of 79.0
To gather data from this node enter the following line into the 'nodes' property above:
```
{field_name="temp", namespace="3", identifier_type="s", identifier="Temperature"},
```


### Example Output

```
opcua,id=n\=3;s\=Temperature temp=79.0,quality="OK (0x0)" 1597820490000000000

```
