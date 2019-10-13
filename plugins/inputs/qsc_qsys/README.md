# Telegraf Plugin: QSC Q-SYS

### Configuration:

```
[[inputs.qsys]]
  ## Specify the core address and port
  server = "localhost:1710"

  ## If the core is set up with user accounts set the username and PIN to use
  # username = "admin"
  # pin = "1234"

  ## If desired, an array of named controls can be collected
  # named_controls = ["CoreTemperature", "Output1Gain"]
```

### Description
The qsys plugin connects to a [QSC Q-SYS](https://www.qsc.com/systems/products/q-sys-ecosystem/) Core audio processor 
using the [QSYS Remote Control protocol](https://q-syshelp.qsc.com/Index.htm#External_Control/Q-Sys_Remote_Control/QRC.htm).

### Measurements:
* qsys
  * state
  * status
  * (any Named Control)

### Tags:
* All measurements have the following tags:
  * design
  * platform
  * server
  
### Example Output:

Using this configuration:
```
[[inputs.qsys]]
  ## Specify the core address and port
  server = "192.168.1.106:1710"
  named_controls = ["CoreProcessorTemperature", "CoreChassisTemperature","TSCStatus","TSCBacklight"]
  username = "admin"
  pin = "1234"
```
```
qsys,design=apartment,host=monitorHost,platform=Core\ 110f,server=192.168.1.106:1710 CoreChassisTemperature=49,CoreProcessorTemperature=64,TSCBacklight=80,TSCStatus=0,state="Active",status="OK - 4 OK" 1570994404000000000
```


