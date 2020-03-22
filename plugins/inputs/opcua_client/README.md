# Telegraf Input Plugin: opcua_client

The opcua_client plugin retrieves data from OPCUA slave devices

### Configuration:

```toml
 ## Connection Configuration
 ##
 ## The plugin supports connections to PLCs via OPCUA 
 ##
 ## Device name
 name = "Device"
 
 # OPC UA Endpoint URL
 endpoint = "opc.tcp://localhost:4840"
 
 ## Read Timeout
 ## add an arbitrary timeout (seconds) to demonstrate how to stop a subscription
 ## with a context.
 timeout = 30

 # Time Inteval, default = 100 * time.Millisecond
 # interval = "10000000"

 # Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto
 policy = "None"
 
 # Security mode: None, Sign, SignAndEncrypt. Default: auto
 mode = "None"
 
 # Path to cert.pem. Required for security mode/policy != None
 # cert = "/etc/telegraf/cert.pem"
 
 # Path to private key.pem. Required for security mode/policy != None
 # key = "/etc/telegraf/key.pem"

 ## Measurements
 ## node id to subscribe to
 ## name       			- the variable name
 ## namespace  			- integer value 0 thru 3
 ## identifier_type		- s=string, i=numeric, g=guid, b=opaque
 ## identifier			- tag as shown in opcua browser
 ## data_type  			- boolean, byte, short, int, uint, uint16, int16, uint32, int32, float, double, string, datetime, number
 ## Template 			- {name="", namespace="", identifier_type="", identifier="", data_type="", description=""},
 nodes = [
	{name="NameSpaceArray", namespace="0", identifier_type="i", identifier="2255", data_type="string", description="List of NameSpace"},
	{name="State", 		namespace="0", identifier_type="i", identifier="2259", data_type="int32", description="Running Status"},
	{name="CurrentTime", 	namespace="0", identifier_type="i", identifier="2258", data_type="datetime", description="OPC UA Server Time"},
	{name="Sine1", 		namespace="2", identifier_type="s",  identifier="Simulation Examples.Functions.Sine1", data_type="float", description="Simulation Sine Wave"},
	]
```
### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter opcua_client -test

```
