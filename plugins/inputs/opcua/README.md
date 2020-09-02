# Telegraf Input Plugin: opcua_client

The opcua_client plugin retrieves data from OPCUA slave devices

### Configuration:

```toml

# ## Connection Configuration
#  ##
#  ## The plugin supports connections to PLCs via OPCUA
#  ##
#  ## Device name
name = "opcua_rocks"
#
#  # OPC UA Endpoint URL
endpoint = "opc.tcp://opcua.rocks:4840"
#
#  ## Read Timeout
#  ## add an arbitrary timeout (seconds) to demonstrate how to stop a subscription
#  ## with a context.
timeout = 30
#
#  # Time Inteval, default = 10s
time_interval = "5s"
#
#  # Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto
security_policy = "None"
#
#  # Security mode: None, Sign, SignAndEncrypt. Default: auto
security_mode = "None"
#
#  # Path to cert.pem. Required for security mode/policy != None. If cert path is not supplied, self-signed cert and key will be generated.
#  # certificate = "/etc/telegraf/cert.pem"
#
#  # Path to private key.pem. Required for security mode/policy != None. If key path is not supplied, self-signed cert and key will be generated.
#  # private_key = "/etc/telegraf/key.pem"
#
#  # To authenticate using a specific ID, select chosen method from 'Certificate' or 'UserName'. Else use 'Anonymous.' Defaults to 'Anonymous' if not provided.
#  # auth_method = "Anonymous"
#
#  # Required for auth_method = "UserName"
#  # username = "myusername"
#
#  # Required for auth_method = "UserName"
#  # password = "mypassword"
#
#  ## Measurements
#  ## node id to subscribe to
#  ## name       			- the variable name
#  ## namespace  			- integer value 0 thru 3
#  ## identifier_type		- s=string, i=numeric, g=guid, b=opaque
#  ## identifier			- tag as shown in opcua browser
#  ## data_type  			- boolean, byte, short, int, uint, uint16, int16, uint32, int32, float, double, string, datetime, number
#  ## Template 			- {name="", namespace="", identifier_type="", identifier="", data_type="", description=""},
nodes = [
		{name="ProductName", namespace="0", identifier_type="i", identifier="2261", data_type="string", description="open62541 OPC UA Server"},
		{name="ProductUri", namespace="0", identifier_type="i", identifier="2262", data_type="string", description="http://open62541.org"},
		{name="ManufacturerName", namespace="0", identifier_type="i", identifier="2263", data_type="string", description="open62541"},
]

## Guide:
## An OPC UA node ID may resemble: "n=3,s=Temperature"
## In this example, n=3 is indicating the namespace is '3'.
## s=Temperature is indicting that the identifier type is a 'string' and the indentifier value is 'Temperature'
## This temperature node may have a current value of 79.0, which would possibly make the value a 'float'.
## To gather data from this node you would need to enter the following line into 'nodes' property above:
##     {name="SomeLabel", namespace="3", identifier_type="s", identifier="Temperature", data_type="float", description="Some description."},

```
### Example Output:

```
opcua,host=3c70aee0901e,name=Random,type=double Random=0.018158170305814902 1597820490000000000

```
