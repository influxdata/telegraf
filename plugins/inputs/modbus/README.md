# Telegraf Input Plugin: Modbus

This plugin gather read Discrete Inputs, Coils, Input Registers and Holding Registers via Modbus TCP or Modbus RTU.

### Configuration:

```toml
# Description
#TCP
 #type = "TCP"
 #controller="192.168.0.9"
 #port = 502

 #RTU
 type = "RTU"
 controller="/dev/ttyUSB0"
 baudRate = 9600
 dataBits = 8
 parity = "N"
 stopBits = 1
 
 slaveId = 1
 timeout = 1

 [[inputs.modbus.Registers.InputRegisters.Tags]]
   name = "Voltage"
   order ="AB"	
   scale = "/10"
   address = [
    0      
   ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
   name = "Current"
   order ="CDAB"	
   scale = "/1000"
   address = [
    1,
    2
   ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Power"
    order ="CDAB"	
    scale = "/10"
    address = [
     3,
     4      
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Energy"
    order ="CDAB"	
    scale = "/1000"
    address = [
     5,
     6      
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Frecuency"
    order ="AB"	
    scale = "/10"
    address = [
     7
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "PowerFactor"
    order ="AB"	
    scale = "/100"
    address = [
     8
    ]
```
### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter modbus -test
modbus.InputRegisters,host=orangepizero Current=0,Energy=0,Frecuency=60,Power=0,PowerFactor=0,Voltage=123.9000015258789 1554079521000000000
```