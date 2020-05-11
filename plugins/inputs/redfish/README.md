# Redfish Input Plugin

The `redfish` plugin gathers  metrics and status information about CPU temperature, fanspeed, Powersupply, voltage, hostname and Location details(datacenter,placement,rack and room) of Dell hardware servers for which redfish is enabled.
And also metrics like CPU temperature,fanspeed, Powersupply and hostname metrics for HP Hardware server(redfish should be enabled).

Note: Currently this plugin only supports DELL and HP servers.



### Configuration

```toml
[[inputs.redfish]]
## Server OOB-IP
host = "http://192.0.0.1"

## Username,Password for hardware server
basicauthusername = "test"
basicauthpassword = "test"
## Server Vendor(dell or hp)
server= "dell"
## Resource Id for redfish APIs
id="System.Embedded.1"
## Optional TLS Config
# tls_ca = "/etc/telegraf/ca.pem"
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = false

## Amount of time allowed to complete the HTTP request
# timeout = "5s"
```

### Metrics for Dell Servers

- cpu_temperatures
	- tags:
		- hostname
		- name
		- oob_ip
		- host
	- Fields:
		- datacenter
		- temperature
		- health
		- rack
		- room
		- row
		- state

- fans
	- tags:
		- hostname
		- name
		- oob_ip
		- host
	- Fields:
		- datacenter
		- fanspeed
		- health
		- rack
		- room
		- row
		- state

- voltages
	- tags:
		- hostname
		- name
		- oob_ip
		- host
	- Fields:
		- datacenter
		- voltage
		- health
		- rack
		- room
		- row
		- state

- Powersupply 
	- tags:
		- hostname
		- name
		- oobip
		- host
	- Fields:
		- datacenter
		- health
		- power_capacity_watts
		- power_input_watts
		- power_output_watts
		- rack
		- room
		- row
		- state

### Metrics for HP Servers

- cpu_temperature
	- tags:
		- hostname
		- name
		- oob_ip
		- host
	- Fields:
		- temperature
		- health
		- state

- fans
	- tags:
		- hostname
		- name
		- oob_ip
		- host
	- Fields:
		- fanspeed
		- health
		- state
- Powersupply 
	- tags:
		- hostname
		- name
		- oob_ip
		- host
		- member_id
	- Fields:
		- power_capacity_watts
		- last_powerOutput_watts 
		- line_input_voltage 

### Example Output For HP
```
cpu_temperature,hostname=tpa_hostname,name=01-Inlet\ Ambient,oob_ip=http://127.0.0.1,host=tpa_po health="OK",state="Enabled",temperature="19" 1582612210000000000
cpu_temperature,hostname=tpa_hostname,name=02-CPU\ 1,oob_ip=http://127.0.0.1,host=tpa_po health="OK",state="Enabled",temperature="40" 1582612210000000000
fans,hostname=tpa_hostname,name=Fan\ 4,oob_ip=http://127.0.0.1,host=tpa_po fanspeed="23",health="OK",state="Enabled" 1582612210000000000
fans,hostname=tpa_hostname,name=Fan\ 5,oob_ip=http://127.0.0.1,host=tpa_po fanspeed="23",health="OK",state="Enabled" 1582612210000000000
fans,hostname=tpa_hostname,name=Fan\ 6,oob_ip=http://127.0.0.1,host=tpa_po fanspeed="23",health="OK",state="Enabled" 1582612210000000000
fans,hostname=tpa_hostname,name=Fan\ 7,oob_ip=http://127.0.0.1,host=tpa_po fanspeed="23",health="OK",state="Enabled" 1582612210000000000
powersupply,hostname=tpa_hostname,member_id=0,name=HpeServerPowerSupply,oob_ip=http://127.0.0.1,host=tpa_po last_power_output_watts="109",line_input_voltage="206",power_capacity_watts="800" 1582612210000000000
powersupply,hostname=tpa_hostname,member_id=1,name=HpeServerPowerSupply,oob_ip=http://127.0.0.1,host=tpa_po last_power_output_watts="98",line_input_voltage="204",power_capacity_watts="800" 1582612210000000000

```

### Example Output For Dell
```
cpu_temperature,hostname=test-hostname,name=CPU1\ Temp,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",temperature="41" 1582114112000000000
cpu_temperature,hostname=test-hostname,name=CPU2\ Temp,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",temperature="51" 1582114112000000000
cpu_temperature,hostname=test-hostname,name=System\ Board\ Inlet\ Temp,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",temperature="23" 1582114112000000000
cpu_temperature,hostname=test-hostname,name=System\ Board\ Exhaust\ Temp,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",temperature="33" 1582114112000000000
fans,hostname=test-hostname,name=System\ Board\ Fan1A,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",fanspeed="17760",health="OK",rack="12",room="tbc",row="3",state="Enabled" 1582114112000000000
fans,hostname=test-hostname,name=System\ Board\ Fan1B,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",fanspeed="15360",health="OK",rack="12",room="tbc",row="3",state="Enabled" 1582114112000000000
fans,hostname=test-hostname,name=System\ Board\ Fan2A,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",fanspeed="17880",health="OK",rack="12",room="tbc",row="3",state="Enabled" 1582114112000000000
powersupply,hostname=test-hostname,name=PS1\ Status,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",power_capacity_watts="750",power_input_watts="900",power_output_watts="208",rack="12",room="tbc",row="3",state="Enabled" 1582114112000000000
powersupply,hostname=test-hostname,name=PS2\ Status,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",power_capacity_watts="750",power_input_watts="900",power_output_watts="194",rack="12",room="tbc",row="3",state="Enabled" 1582114112000000000
voltages,hostname=test-hostname,name=CPU1\ MEM345\ VDDQ\ PG,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",voltage="1" 1582114112000000000
voltages,hostname=test-hostname,name=CPU1\ MEM345\ VPP\ PG,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",voltage="1" 1582114112000000000
voltages,hostname=test-hostname,name=CPU1\ MEM345\ VTT\ PG,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",voltage="1" 1582114112000000000
voltages,hostname=test-hostname,name=PS1\ voltage\ 1,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",Voltage="208" 1582114112000000000
voltages,hostname=test-hostname,name=PS2\ voltage\ 2,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",Voltage="208" 1582114112000000000
voltages,hostname=test-hostname,name=System\ Board\ 3.3V\ A\ PG,oob_ip=http://190.0.0.1,host=test-telegraf datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",voltage="1" 1582114112000000000

```
