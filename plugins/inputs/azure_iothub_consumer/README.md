# Telegraf Input Plugin: Azure Iot Hub Consumer

Input plugin for Azure IoT Hub Edge Module

### Configuration:

```toml
## One of the following sets required for configuration:
#  
#  # 1.
#  connection_string = ""
#  use_gateway = true
#
#  # 2.
#  hub_name = ""
#  device_id = ""
#  module_id = ""
#  shared_access_key = ""
#  use_gateway = true
#
#  # 3. (Recommended)
#  Provide no configuration for IoT Edge module, and it will self-configure from environment variables present in edge modules.

connection_string = "HostName=[MYIOTHUB].azure-devices.net;DeviceId=[MYEDGEDEVICE];ModuleId=[MYTELEGRAFMODULE];SharedAccessKey=[MYSHAREDACCESSKEY(Primary)]"
use_gateway = true
```