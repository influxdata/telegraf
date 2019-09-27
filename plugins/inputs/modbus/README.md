# Telegraf Input Plugin: Modbus

The Modbus plugin collects Discrete Inputs, Coils, Input Registers and Holding Registers via Modbus TCP or Modbus RTU/ASCII

### Configuration:

```toml
 slave_id = 1
 timeout = "1s"
 #transmission_mode = "RTU"
 
 #TCP 
 controller = "tcp://localhost:1502"
 
 #RTU
 #controller = "file:///dev/ttyUSB0"
 #baud_rate = 9600
 #data_bits = 8
 #parity = "N"
 #stop_bits = 1
 
 ## Digital Variables, Discrete Inputs and Coils
 ## name    - the variable name
 ## address - variable address
 
 discrete_inputs = [
   { name = "Start",          address = [0]},   
   { name = "Stop",           address = [1]},   
   { name = "Reset",          address = [2]},   
   { name = "EmergencyStop",  address = [3]},   
 ]
 coils = [
   { name = "Motor1-Run",     address = [0]},   
   { name = "Motor1-Jog",     address = [1]},   
   { name = "Motor1-Stop",    address = [2]},      
 ] 
 
 ## Analog Variables, Input Registers and Holding Registers
 ## name       - the variable name 
 ## byte_order - the ordering of bytes 
 ##  |---AB, ABCD   - Big Endian
 ##  |---BA, DCBA   - Little Endian
 ##  |---BADC       - Mid-Big Endian
 ##  |---CDAB       - Mid-Little Endian
 ## data_type  - UINT16, INT16, INT32, UINT32, FLOAT32, FLOAT32-IEEE (the IEEE 754 binary representation)
 ## scale      - the final numeric variable representation    
 ## address    - variable address
 
 holding_registers = [
   { name = "PowerFactor", byte_order = "AB",   data_type = "FLOAT32", scale="0.01" ,  address = [8]},
   { name = "Voltage",     byte_order = "AB",   data_type = "FLOAT32", scale="0.1" ,   address = [0]},   
   { name = "Energy",      byte_order = "ABCD", data_type = "FLOAT32", scale="0.001" , address = [5,6]},
   { name = "Current",     byte_order = "ABCD", data_type = "FLOAT32", scale="0.001" , address = [1, 2]},
   { name = "Frequency",   byte_order = "AB",   data_type = "FLOAT32", scale="0.1" ,   address = [7]},
   { name = "Power",       byte_order = "ABCD", data_type = "FLOAT32", scale="0.1" ,   address = [3,4]},      
 ] 
 input_registers = [
   { name = "TankLevel",   byte_order = "AB",   data_type = "INT16",   scale="1" ,     address = [0]},
   { name = "TankPH",      byte_order = "AB",   data_type = "INT16",  scale="1" ,     address = [1]},   
   { name = "Pump1-Speed", byte_order = "ABCD", data_type = "INT32",   scale="1" ,     address = [3,4]},
 ]
```
### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter modbus -test
modbus.InputRegisters,host=orangepizero Current=0,Energy=0,Frecuency=60,Power=0,PowerFactor=0,Voltage=123.9000015258789 1554079521000000000
```
