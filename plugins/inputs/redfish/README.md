# Redfish Input Plugin

The `redfish` plugin gathers  metrics and status information about CPU temperature, fanspeed, Powersupply, voltage, hostname and Location details(datacenter,placement,rack and room) of hardware servers for which redfish is enabled.


### Configuration

```toml
[[inputs.redfish]]
  ## Server OOB-IP
  address = "http://192.0.0.1"

  ## Username,Password for hardware server
  username = "test"
  password = "test"

  ## Resource Id for redfish APIs
  id="System.Embedded.1"

  ## Optional TLS Config 
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  # insecure_skip_verify = false

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"
```

### Metrics 

- redfish_power_powersupplies
        - tags:
                - address
                - name
                - datacenter (available only if location data is found)
                - rack (available only if location data is found)
                - room (available only if location data is found)
                - row (available only if location data is found)
                - state
                - health
        - Fields:
                - last_power_output_watts
                - line_input_voltage
                - power_capacity_watts
                - power_input_watts
                - power_output_watts

- redfish_power_voltages (available only if voltage data is found)
        - tags:
                - address
                - name
                - datacenter (available only if location data is found)
                - rack (available only if location data is found)
                - room (available only if location data is found)
                - row (available only if location data is found)
                - state
                - health
        - Fields:
                - reading_volts
                - upper_threshold_critical
                - upper_threshold_fatal


- redfish_thermal_fans
        - tags:
                - address
                - name
                - datacenter (available only if location data is found)
                - rack (available only if location data is found)
                - room (available only if location data is found)
                - row (available only if location data is found)
                - state
                - health
        - Fields:
                - reading_rpm (or) reading_percent
                - upper_threshold_critical
                - upper_threshold_fatal


- redfish_thermal_temperatures
        - tags:
                - address
                - name
                - datacenter (available only if location data is found)
                - rack (available only if location data is found)
                - room (available only if location data is found)
                - row (available only if location data is found)
                - state
                - health
        - Fields:
                - reading_celsius
                - upper_threshold_critical
                - upper_threshold_fatal



### Example Output 
```
redfish_thermal_temperatures,source=test-hostname,name=CPU1\ Temp,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_celsius=41,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_temperatures,source=test-hostname,name=CPU2\ Temp,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_celsius=51,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_temperatures,source=test-hostname,name=System\ Board\ Inlet\ Temp,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_celsius=23,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_temperatures,source=test-hostname,name=System\ Board\ Exhaust\ Temp,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_celsius=33,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_fans,source=test-hostname,name=System\ Board\ Fan1A,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading=17720,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_fans,source=test-hostname,name=System\ Board\ Fan1B,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading=17760,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_thermal_fans,source=test-hostname,name=System\ Board\ Fan2A,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading=17880,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_powersupplies,source=test-hostname,name=PS1\ Status,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" power_capacity_watts=750,power_input_watts=900,power_output_watts=208,last_power_output_watts=98,line_input_reading_volts=204 1582114112000000000
redfish_power_powersupplies,source=test-hostname,name=PS2\ Status,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" power_capacity_watts=750,power_input_watts=900,power_output_watts=194,last_power_output_watts=98,line_input_reading_volts=204 1582114112000000000
redfish_power_voltages,source=test-hostname,name=CPU1\ MEM345\ VDDQ\ PG,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_volts=1,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=CPU1\ MEM345\ VPP\ PG,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_volts=1,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=CPU1\ MEM345\ VTT\ PG,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_volts=1,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=PS1\ voltage\ 1,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_volts=208,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=PS2\ voltage\ 2,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_volts=208,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000
redfish_power_voltages,source=test-hostname,name=System\ Board\ 3.3V\ A\ PG,address=http://190.0.0.1,host=test-telegraf,datacenter="Tampa",health="OK",rack="12",room="tbc",row="3",state="Enabled" reading_volts=1,upper_threshold_critical=59,upper_threshold_fatal=64 1582114112000000000


```
