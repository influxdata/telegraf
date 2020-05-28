# Telegraf Output Plugin: Azure Iot Hub

Output plugin for Azure IoT Hub Edge Module

### Configuration:

```toml
## One of the following sets required for configuration:
#  
## For use on IoT Edge: (creates client from environment variables)
#
# use_gateway = true
#
## To specify a device/module connection string:
#
#  connection_string = ""
#  use_gateway = true
#
## To use a shared access key to form a connection string
#
#  hub_name = ""
#  device_id = ""
#  module_id = ""
#  shared_access_key = ""
#  use_gateway = true

connection_string = "HostName=[MYIOTHUB].azure-devices.net;DeviceId=[MYEDGEDEVICE];ModuleId=[MYTELEGRAFMODULE];SharedAccessKey=[MYSHAREDACCESSKEY(Primary)]"
use_gateway = true
```