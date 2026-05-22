# Redfish Input Plugin

This plugin gathers metrics and status information of server hardware with
enabled [DMTF's Redfish][redfish] support.

⭐ Telegraf v1.15.0
🏷️ server
💻 all

[redfish]: https://redfish.dmtf.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret store support

This plugin supports secrets from secret stores for the `username` and
`password` options. See the [secret store documentation][SECRETSTORE] for more
details on how to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Read CPU, Fans, Powersupply and Voltage metrics of hardware server through redfish APIs
[[inputs.redfish]]
  ## Redfish API Base URL.
  address = "https://127.0.0.1:5000"

  ## Credentials for the Redfish API. Use either a username and password or a Token. Can also use secrets.
  username = "root"
  password = "password123456"
  # token = "0123456789abcdef"

  ## System Id to collect data for in Redfish APIs.
  ## Examples: Dell: System.Embedded.1 HPE: 1
  computer_system_id="System.Embedded.1"

  ## Metrics to collect
  ## The metric collects to gather. Choose from "power", "thermal" and "storage".
  # include_metrics = ["power", "thermal", "storage"]

  ## Tag sets allow you to include redfish OData link parent data
  ## For Example.
  ## Thermal data is an OData link with parent Chassis which has a link of Location.
  ## For more info see the Redfish Resource and Schema Guide at DMTFs website.
  ## Available sets are: "chassis.location" and "chassis"
  # include_tag_sets = ["chassis.location"]

  ## Workarounds
  ## Defines workarounds for certain hardware vendors. Choose from:
  ## * ilo4-thermal - Do not pass 0Data-Version header to Thermal endpoint
  ##   deprecated in 1.39; ILO4 is EOSL, option will be ignored
  # workarounds = []

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Metrics
There are two types of metrics. The newer subsystem type and the old type.

- redfish_thermal_temperatures
  - tags:
    - source
    - member_id
    - address
    - name
    - state
    - health
  - fields:
    - reading_celsius
    - upper_threshold_critical
    - upper_threshold_fatal
    - lower_threshold_critical
    - lower_threshold_fatal

- redfish_thermal_fans
  - tags:
    - source
    - member_id
    - address
    - name
    - state
    - health
  - fields:
    - reading_rpm (or) reading_percent
    - upper_threshold_critical
    - upper_threshold_fatal
    - lower_threshold_critical
    - lower_threshold_fatal

- redfish_thermalsubsys_temperatures
  - tags
    - name
    - source
    - address
    - state
    - health_rollup
  - fields
    - reading_celsius

- redfish_thermalsubsys_fans
  - tags
    - member_id
    - name
    - address
    - source
    - state
    - health
  - fields
    - upper_threshold_critical
    - upper_threshold_fatal
    - lower_threshold_critical
    - lower_threshold_fatal
    - reading_rpm (or) reading_percent

- redfish_power_powercontrol
  - tags
		- member_id
	  - address
    - name
		- source
  - fields
			- power_allocated_watts
			- power_available_watts
			- power_capacity_watts
			- power_consumed_watts
			- power_requested_watts
			- average_consumed_watts
			- interval_in_min
			- max_consumed_watts
			- min_consumed_watts

- redfish_power_powersupplies
  - tags:
    - source
    - member_id
    - address
    - name
    - state
    - health
    - serial_num
  - fields:
    - last_power_output_watts
    - line_input_voltage
    - power_capacity_watts
    - power_input_watts
    - power_output_watts

- redfish_power_voltages (available only if voltage data is found)
  - tags:
    - source
    - member_id
    - address
    - name
    - state
    - health
  - fields:
    - reading_volts
    - upper_threshold_critical
    - upper_threshold_fatal
    - lower_threshold_critical
    - lower_threshold_fatal

- redfish_powersubsys_redundancy
	-tags
    - name
		- address
		- source
		- type
		- health
		- state
  - fields
		- redund_group_count

- redfish_powersubsys_powersupplies
	- tags
		- address
		- name
		- source
		- state
		- serial_num
		- hotpluggable
		- health
  - fields
		- power_input_watts
		- power_output_watts
		- line_input_voltage
		- power_capacity_watts
		- firmware_version

- redfish_storage
  - tags
		- source
		- address
		- state
		- health_rollup
		- manufacturer
		- media_type
		- model
		- location
		- protocol
		- serial_number
		- disk_health
		- disk_state
	- fields
    - speed_gbs
		- capacity_bytes

### Tag Sets

- chassis.location
  - tags:
    - datacenter (deprecated in 1.39 since its not part of the redfish standard)
    - rack (available only if location data is found)
    - room (available only if location data is found)
    - row (available only if location data is found)

- chassis
  - tags:
    - chassis_chassistype
    - chassis_manufacturer
    - chassis_model
    - chassis_partnumber
    - chassis_powerstate
    - chassis_sku
    - chassis_serialnumber
    - chassis_state
    - chassis_health

## Example Output

```text
redfish_thermal_temperatures,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=OK,member_id=0,name=CPU1\ Temp,rack=WEB43,row=North,source=web483,state=Enabled reading_celsius=41,upper_threshold_critical=45,upper_threshold_fatal=48 1691270160000000000
redfish_thermal_temperatures,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,member_id=1,name=CPU2\ Temp,rack=WEB43,row=North,source=web483,state=Disabled upper_threshold_critical=45,upper_threshold_fatal=48 1691270160000000000
redfish_thermal_temperatures,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=OK,member_id=2,name=Chassis\ Intake\ Temp,rack=WEB43,row=North,source=web483,state=Enabled upper_threshold_critical=40,upper_threshold_fatal=50,lower_threshold_critical=5,lower_threshold_fatal=0,reading_celsius=25 1691270160000000000
redfish_thermal_fans,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=OK,member_id=0,name=BaseBoard\ System\ Fan,rack=WEB43,row=North,source=web483,state=Enabled lower_threshold_fatal=0i,reading_rpm=2100i 1691270160000000000
redfish_thermal_fans,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=OK,member_id=1,name=BaseBoard\ System\ Fan\ Backup,rack=WEB43,row=North,source=web483,state=Enabled lower_threshold_fatal=0i,reading_rpm=2050i 1691270160000000000
redfish_power_powersupplies,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=Warning,member_id=0,name=Power\ Supply\ Bay,rack=WEB43,row=North,source=web483,state=Enabled line_input_voltage=120,last_power_output_watts=325,power_capacity_watts=800 1691270160000000000
redfish_power_voltages,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=OK,member_id=0,name=VRM1\ Voltage,rack=WEB43,row=North,source=web483,state=Enabled upper_threshold_fatal=15,lower_threshold_critical=11,lower_threshold_fatal=10,reading_volts=12,upper_threshold_critical=13 1691270160000000000
redfish_power_voltages,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=OK,member_id=1,name=VRM2\ Voltage,rack=WEB43,row=North,source=web483,state=Enabled reading_volts=5,upper_threshold_critical=7,lower_threshold_critical=4.5 1691270160000000000
redfish_thermal_temperatures,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=OK,member_id=0,name=CPU1\ Temp,rack=WEB43,row=North,source=web483,state=Enabled upper_threshold_critical=45,upper_threshold_fatal=48,reading_celsius=41 1691270170000000000
redfish_thermal_temperatures,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,member_id=1,name=CPU2\ Temp,rack=WEB43,row=North,source=web483,state=Disabled upper_threshold_critical=45,upper_threshold_fatal=48 1691270170000000000
redfish_thermal_temperatures,address=127.0.0.1,chassis_chassistype=RackMount,chassis_health=OK,chassis_manufacturer=Contoso,chassis_model=3500RX,chassis_partnumber=224071-J23,chassis_powerstate=On,chassis_serialnumber=437XR1138R2,chassis_sku=8675309,chassis_state=Enabled,health=OK,member_id=2,name=Chassis\ Intake\ Temp,rack=WEB43,row=North,source=web483,state=Enabled lower_threshold_critical=5,lower_threshold_fatal=0,reading_celsius=25,upper_threshold_critical=40,upper_threshold_fatal=50 1691270170000000000
redfish_storage,address=127.0.0.1,disk_health=Critical,disk_state=UnavailableOffline,health_rollup=Critical,host=testhost,location=Slot\=1:Port\=1I:Box\=1:Bay\=4,media_type=HDD,model=ST2000LOLOLOL,protocol=SATA,serial_number=ICT1200,source=testhost,state=Enabled capacity_bytes=2000398934016i,speed_gbs=6 1779446190000000000
```
