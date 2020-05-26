# Redfish Input Plugin

The `redfish` plugin gathers  metrics and status information about CPU temperature, fanspeed, Powersupply, voltage, hostname and Location details(datacenter,placement,rack and room) of hardware servers for which redfish is enabled.


### Configuration

```toml
[[inputs.redfish]]
  ## Server OOB-IP
  host = "192.0.0.1"

  ## Username,Password for hardware server
  username = "test"
  password = "test"

  ## Resource Id for redfish APIs
  id="System.Embedded.1"

  ## Optional TLS Config, if not provided insecure skip verifies defaults to true 
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"
```

### Metrics for Dell Servers

- redfish_power_powersupplies
        - tags:
                - source_ip
                - name
                - datacenter
                - rack
                - room
                - row
                - state
                - health
        - Fields:
                - last_power_output_watts
                - line_input_voltage
                - power_capacity_watts
                - power_input_watts
                - power_output_watts

- redfish_power_voltages
        - tags:
                - source_ip
                - name
                - datacenter
                - rack
                - room
                - row
                - state
                - health
		- severity
        - Fields:
                - voltage
                - upper_threshold_critical
                - upper_threshold_fatal


- redfish_thermal_fans
        - tags:
                - source_ip
                - name
                - datacenter
                - rack
                - room
                - row
                - state
                - health
		- severity
        - Fields:
                - fanspeed
                - upper_threshold_critical
                - upper_threshold_fatal


- redfish_thermal_temperatures
        - tags:
                - source_ip
                - name
                - datacenter
                - rack
                - room
                - row
                - state
                - health
		- severity
        - Fields:
                - temperature
                - upper_threshold_critical
                - upper_threshold_fatal



### Metrics if location details, voltage and power input/output data are not available in server APIs

- redfish_power_powersupplies
        - tags:
                - health
                - host
                - name
                - source_ip
                - state
        - Fields:
                - last_power_output_watts
                - line_input_voltage
                - power_capacity_watts

- redfish_thermal_fans
        - tags:
                - health
                - host
                - name
                - source_ip
                - state
		- severity
        - Fields:
                - fanspeed
		- upper_threshold_critical
		- upper_threshold_fatal

- redfish_thermal_temperatures
        - tags:
                - health
                - host
                - name
                - source_ip
                - state
		- severity
        - Fields:
                - temperature
		- upper_threshold_critical
		- upper_threshold_fatal




### Example Output For Dell
```
redfish_thermal_temperatures,source=test-hostname,name=CPU1\ Temp,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" temperature=41,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_temperatures,source=test-hostname,name=CPU2\ Temp,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" temperature=51,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_temperatures,source=test-hostname,name=System\ Board\ Inlet\ Temp,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" temperature=23,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_temperatures,source=test-hostname,name=System\ Board\ Exhaust\ Temp,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" temperature=33,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_fans,source=test-hostname,name=System\ Board\ Fan1A,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" fanspeed=17720,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_fans,source=test-hostname,name=System\ Board\ Fan1B,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" fanspeed=17760,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_fans,source=test-hostname,name=System\ Board\ Fan2A,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" fanspeed=17880,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_powersupplies,source=test-hostname,name=PS1\ Status,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" power_capacity_watts=750,power_input_watts=900,power_output_watts=208,last_power_output_watts=98,line_input_voltage=204 1582114112000000000
redfish_power_powersupplies,source=test-hostname,name=PS2\ Status,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" power_capacity_watts=750,power_input_watts=900,power_output_watts=194,last_power_output_watts=98,line_input_voltage=204 1582114112000000000
redfish_power_voltages,source=test-hostname,name=CPU1\ MEM345\ VDDQ\ PG,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" voltage=1,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=CPU1\ MEM345\ VPP\ PG,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" voltage=1,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=CPU1\ MEM345\ VTT\ PG,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" voltage=1,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=PS1\ voltage\ 1,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" voltage=208,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=PS2\ voltage\ 2,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" voltage=208,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=System\ Board\ 3.3V\ A\ PG,source_ip=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled",severity="OK" voltage=1,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000


```


### Example output if location details, voltage and power input/output data are not available in server APIs

```
redfish_thermal_temperatures,source=tpa_hostname,name=01-Inlet\ Ambient,source_ip=http://127.0.0.1,health="OK",state="Enabled",severity="OK" temperature=19,upper_threshold_critical=59,upper_threshold_fatal=64 1582612210000000000
redfish_thermal_temperatures,source=tpa_hostname,name=02-CPU\ 1,source_ip=http://127.0.0.1,,health="OK",state="Enabled",severity="OK" temperature=40,upper_threshold_critical=59,upper_threshold_fatal=64 1582612210000000000
redfish_thermal_fans,source=tpa_hostname,name=Fan\ 4,source_ip=http://127.0.0.1,health="OK",state="Enabled",severity="OK" fanspeed=23,upper_threshold_critical=59,upper_threshold_fatal=64 1582612210000000000
redfish_thermal_fans,source=tpa_hostname,name=Fan\ 5,source_ip=http://127.0.0.1,health="OK",state="Enabled",severity="OK" fanspeed=28,upper_threshold_critical=59,upper_threshold_fatal=64 1582612210000000000
redfish_power_powersupplies,source=tpa_hostname,name=HpeServerPowerSupply,source_ip=http://127.0.0.1,health="OK",state="Enabled" last_power_output_watts="109",line_input_voltage="206",power_capacity_watts="800" 1582612210000000000
redfish_power_powersupplies,source=tpa_hostname,name=HpeServerPowerSupply,source_ip=http://127.0.0.1,health="OK",state="Enabled"  last_power_output_watts="98",line_input_voltage="204",power_capacity_watts="800" 1582612210000000000

```
