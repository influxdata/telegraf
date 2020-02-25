# Redfish Input Plugin

The `redfish` plugin gathers  metrics and status information about CPU temperature, Fanspeed, Powersupply, voltage, Hostname and Location details(datacenter,placement,rack and room) of Dell hardware servers for which redfish is enabled.
And also metrics like CPU temperature,Fanspeed, Powersupply and Hostname metrics for HP Hardware server(redfish should be enabled).

Note: Currently this plugin only supports DELL and HP servers.



### Configuration

```toml
[[inputs.redfish]]
## Server OOB-IP
host = "https://192.0.0.1"

## Username,Password for hardware server
basicauthusername = "test"
basicauthpassword = "test"
## Server Vendor(dell or hp)
server= "dell"
## Resource Id for redfish APIs
id="System.Embedded.1"

## Amount of time allowed to complete the HTTP request
# timeout = "5s"
```

### Metrics for Dell Servers

- cputemperatures
	- tags:
		- Hostname
		- Name
		- OOBIP
		- host
	- Fields:
		- Datacenter
		- Temperature
		- Health
		- Rack
		- Room
		- Row
		- State

- fans
	- tags:
		- Hostname
		- Name
		- OOBIP
		- host
	- Fields:
		- Datacenter
		- Fanspeed
		- Health
		- Rack
		- Room
		- Row
		- State

- Voltages
	- tags:
		- Hostname
		- Name
		- OOBIP
		- host
	- Fields:
		- Datacenter
		- Voltage
		- Health
		- Rack
		- Room
		- Row
		- State

- Powersupply 
	- tags:
		- Hostname
		- Name
		- OOBIP
		- host
	- Fields:
		- Datacenter
		- Health
		- PowerCapacityWatts
		- PowerInputWatts
		- PowerOutputWatts
		- Rack
		- Room
		- Row
		- State

### Metrics for HP Servers

- cputemperatures
	- tags:
		- Hostname
		- Name
		- OOBIP
		- host
	- Fields:
		- Temperature
		- Health
		- State

- fans
	- tags:
		- Hostname
		- Name
		- OOBIP
		- host
	- Fields:
		- Fanspeed
		- Health
		- State
- Powersupply 
	- tags:
		- Hostname
		- Name
		- OOBIP
		- host
		- MemberId
	- Fields:
		- PowerCapacityWatts
		- LastPowerOutputWatts 
		- LineInputVoltage 

### Example Output For HP
```
cputemperature,Hostname=tpa_hostname,Name=01-Inlet\ Ambient,OOBIP=https://127.0.0.1,host=tpa_po Health="OK",State="Enabled",Temperature="19" 1582612210000000000
cputemperature,Hostname=tpa_hostname,Name=02-CPU\ 1,OOBIP=https://127.0.0.1,host=tpa_po Health="OK",State="Enabled",Temperature="40" 1582612210000000000
fans,Hostname=tpa_hostname,Name=Fan\ 4,OOBIP=https://127.0.0.1,host=tpa_po Fanspeed="23",Health="OK",State="Enabled" 1582612210000000000
fans,Hostname=tpa_hostname,Name=Fan\ 5,OOBIP=https://127.0.0.1,host=tpa_po Fanspeed="23",Health="OK",State="Enabled" 1582612210000000000
fans,Hostname=tpa_hostname,Name=Fan\ 6,OOBIP=https://127.0.0.1,host=tpa_po Fanspeed="23",Health="OK",State="Enabled" 1582612210000000000
fans,Hostname=tpa_hostname,Name=Fan\ 7,OOBIP=https://127.0.0.1,host=tpa_po Fanspeed="23",Health="OK",State="Enabled" 1582612210000000000
powersupply,Hostname=tpa_hostname,MemberId=0,Name=HpeServerPowerSupply,OOBIP=https://127.0.0.1,host=tpa_po LastPowerOutputWatts="109",LineInputVoltage="206",PowerCapacityWatts="800" 1582612210000000000
powersupply,Hostname=tpa_hostname,MemberId=1,Name=HpeServerPowerSupply,OOBIP=https://127.0.0.1,host=tpa_po LastPowerOutputWatts="98",LineInputVoltage="204",PowerCapacityWatts="800" 1582612210000000000

```

### Example Output For Dell
```
cputemperature,Hostname=test-hostname,Name=CPU1\ Temp,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Temperature="41" 1582114112000000000
cputemperature,Hostname=test-hostname,Name=CPU2\ Temp,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Temperature="51" 1582114112000000000
cputemperature,Hostname=test-hostname,Name=System\ Board\ Inlet\ Temp,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Temperature="23" 1582114112000000000
cputemperature,Hostname=test-hostname,Name=System\ Board\ Exhaust\ Temp,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Temperature="33" 1582114112000000000
fans,Hostname=test-hostname,Name=System\ Board\ Fan1A,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Fanspeed="17760",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled" 1582114112000000000
fans,Hostname=test-hostname,Name=System\ Board\ Fan1B,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Fanspeed="15360",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled" 1582114112000000000
fans,Hostname=test-hostname,Name=System\ Board\ Fan2A,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Fanspeed="17880",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled" 1582114112000000000
powersupply,Hostname=test-hostname,Name=PS1\ Status,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",PowerCapacityWatts="750",PowerInputWatts="900",PowerOutputWatts="208",Rack="12",Room="tbc",Row="3",State="Enabled" 1582114112000000000
powersupply,Hostname=test-hostname,Name=PS2\ Status,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",PowerCapacityWatts="750",PowerInputWatts="900",PowerOutputWatts="194",Rack="12",Room="tbc",Row="3",State="Enabled" 1582114112000000000
voltages,Hostname=test-hostname,Name=CPU1\ MEM345\ VDDQ\ PG,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Voltage="1" 1582114112000000000
voltages,Hostname=test-hostname,Name=CPU1\ MEM345\ VPP\ PG,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Voltage="1" 1582114112000000000
voltages,Hostname=test-hostname,Name=CPU1\ MEM345\ VTT\ PG,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Voltage="1" 1582114112000000000
voltages,Hostname=test-hostname,Name=PS1\ Voltage\ 1,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Voltage="208" 1582114112000000000
voltages,Hostname=test-hostname,Name=PS2\ Voltage\ 2,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Voltage="208" 1582114112000000000
voltages,Hostname=test-hostname,Name=System\ Board\ 3.3V\ A\ PG,OOBIP=https://190.0.0.1,host=test-telegraf Datacenter="Tampa",Health="OK",Rack="12",Room="tbc",Row="3",State="Enabled",Voltage="1" 1582114112000000000

```
