# neptune_apex Input Plugin

The neptune_apex Input Plugin collects real-time data from the Apex's status.xml page.

### Configuration

```toml
[[inputs.neptune_apex]]
  ## The Neptune Apex plugin reads the publicly available status.xml data from a local Apex.
  ## Measurements will be logged under "apex".

  ## The hostname/IP of the local Apex(es). If you specify more than one server, they will
  ## be differentiated by the "hostname" tag.
  servers = [
    "apex.local",
  ]

  ## The response_timeout specifies how long to wait for a reply from the Apex.
  #response_timeout = "5s"

```

### Metrics

The [Neptune Apex](https://www.neptunesystems.com/) controller family allows an aquarium hobbyist to monitor and control
their tanks based on various probes. The data is taken directly from the /cgi-bin/status.xml at the interval specified 
in the telegraf.conf configuration file. 

No manipulation is done on any of the fields to ensure future changes to the status.xml do not introduce conversion bugs
to this plugin. When reasonable and predictable, some tags are derived to make graphing easier and without front-end 
programming. These tags are clearly marked in the list below and should be considered a convenience rather than authoritative.

- neptune_apex (All metrics have this measurement name)
  - tags:
    - host (mandatory, string) is the host on which telegraf runs.
    - hostname (mandatory, string) contains the hostname of the apex device. This can be used to differentiate between 
    different units. By using the hostname instead of the serial number, replacements units won't disturb graphs.
    - type (mandatory, string) maps to the different types of data. Values can be "controller" (The Apex controller 
    itself), "probe" for the different input probes, or "output" for any physical or virtual outputs. The Watt and Amp 
    probes attached to the physical 120V outlets are aggregated under the output type.
    - probe_type (optional, string) contains the probe type as reported by the Apex.
    - name (optional, string) contains the name of the probe or output.
    - output_id (optional, string) represents the internal unique output ID. This is different from the device_id.
    - device_id (optional, string) maps to either the aquabus address or the internal reference.
    - output_type (optional, string) categorizes the output into different categories. This tag is DERIVED from the 
    device_id. Possible values are: "variable" for the 0-10V signal ports, "outlet" for physical 120V sockets, "alert" 
    for alarms (email, sound), "virtual" for user-defined outputs, and "unknown" for everything else.
  - fields:
    - value (float, various unit) represents the probe reading.
    - state (string) represents the output state as defined by the Apex. Examples include "AOF" for Auto (OFF), "TBL" 
    for operating according to a table, and "PF*" for different programs.
    - amp (float, Ampere) is the amount of current flowing through the 120V outlet.
    - watt (float, Watt) represents the amount of energy flowing through the 120V outlet.
    - xstatus (string) indicates the xstatus of an outlet. Found on wireless Vortech devices.
    - hardware (string, controller hardware version)
    - software (string, software version)
    - power_failed (string, date time) when the controller last lost power.
    - power_restored (string, date time) when the controller last powered on.
    - serial (string, serial number)
    - timezone (float, timezone offset)
   - time:
     - The time used for the metric is parsed from the status.xml page. This helps when cross-referencing events with 
     the local system of Apex Fusion. Since the Apex uses NTP, this should not matter in most scenarios. 
     

### Sample Queries


Get the max, mean, and min for the temperature in the last hour:
```
SELECT mean("value") FROM "neptune_apex" WHERE ("probe_type" = 'Temp') AND time >= now() - 6h GROUP BY time(20s)
```

### Troubleshooting

#### sendRequest failure
This indicates a problem communicating with the local Apex controller. If on Mac/Linux, try curl:
```
$ curl apex.local/cgi-bin/status.xml
```
to isolate the problem.

#### parseXML errors
Ensure the XML being returned is valid. If you get valid XML back, open a bug request.

#### Missing fields/data
The neptune_apex plugin is strict on its input to prevent any conversion errors. If you have fields in the status.xml
output that are not converted to a metric, open a feature request and paste your whole status.xml

### Example Output

```
> neptune_apex,host=ubuntu,hostname=apex hardware="1.0",power_failed="12/15/2018 11:00:00",power_restored="12/15/2018 16:31:15",serial="AC5:12345",software="5.04_7A18",timezone=-8 1545716958000000000
> neptune_apex,device_id=base_Var1,host=ubuntu,hostname=apex,name=VarSpd1_I1,output_id=0,output_type=variable,type=output state="PF1" 1545716958000000000
> neptune_apex,device_id=base_Var2,host=ubuntu,hostname=apex,name=VarSpd2_I2,output_id=1,output_type=variable,type=output state="PF2" 1545716958000000000
> neptune_apex,device_id=base_Var3,host=ubuntu,hostname=apex,name=VarSpd3_I3,output_id=2,output_type=variable,type=output state="PF3" 1545716958000000000
> neptune_apex,device_id=base_Var4,host=ubuntu,hostname=apex,name=VarSpd4_I4,output_id=3,output_type=variable,type=output state="PF4" 1545716958000000000
> neptune_apex,device_id=base_Alarm,host=ubuntu,hostname=apex,name=SndAlm_I6,output_id=4,output_type=alert,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=base_Warn,host=ubuntu,hostname=apex,name=SndWrn_I7,output_id=5,output_type=alert,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=base_email,host=ubuntu,hostname=apex,name=EmailAlm_I5,output_id=6,output_type=alert,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=base_email2,host=ubuntu,hostname=apex,name=Email2Alm_I9,output_id=7,output_type=alert,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=2_1,host=ubuntu,hostname=apex,name=RETURN_2_1,output_id=8,output_type=outlet,type=output amp=0.3,state="AON",watt=33 1545716958000000000
> neptune_apex,device_id=2_2,host=ubuntu,hostname=apex,name=Heater1_2_2,output_id=9,output_type=outlet,type=output amp=1.2,state="AON",watt=139 1545716958000000000
> neptune_apex,device_id=2_3,host=ubuntu,hostname=apex,name=FREE_2_3,output_id=10,output_type=outlet,type=output amp=0,state="OFF",watt=1 1545716958000000000
> neptune_apex,device_id=2_4,host=ubuntu,hostname=apex,name=LIGHT_2_4,output_id=11,output_type=outlet,type=output amp=0,state="OFF",watt=1 1545716958000000000
> neptune_apex,device_id=2_5,host=ubuntu,hostname=apex,name=LHead_2_5,output_id=12,output_type=outlet,type=output amp=0,state="AON",watt=3 1545716958000000000
> neptune_apex,device_id=2_6,host=ubuntu,hostname=apex,name=SKIMMER_2_6,output_id=13,output_type=outlet,type=output amp=0.1,state="AON",watt=11 1545716958000000000
> neptune_apex,device_id=2_7,host=ubuntu,hostname=apex,name=FREE_2_7,output_id=14,output_type=outlet,type=output amp=0,state="OFF",watt=1 1545716958000000000
> neptune_apex,device_id=2_8,host=ubuntu,hostname=apex,name=CABLIGHT_2_8,output_id=15,output_type=outlet,type=output amp=0,state="AON",watt=1 1545716958000000000
> neptune_apex,device_id=2_9,host=ubuntu,hostname=apex,name=LinkA_2_9,output_id=16,output_type=unknown,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=2_10,host=ubuntu,hostname=apex,name=LinkB_2_10,output_id=17,output_type=unknown,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=3_1,host=ubuntu,hostname=apex,name=RVortech_3_1,output_id=18,output_type=unknown,type=output state="TBL",xstatus="OK" 1545716958000000000
> neptune_apex,device_id=3_2,host=ubuntu,hostname=apex,name=LVortech_3_2,output_id=19,output_type=unknown,type=output state="TBL",xstatus="OK" 1545716958000000000
> neptune_apex,device_id=4_1,host=ubuntu,hostname=apex,name=OSMOLATO_4_1,output_id=20,output_type=outlet,type=output amp=0,state="AON",watt=1 1545716958000000000
> neptune_apex,device_id=4_2,host=ubuntu,hostname=apex,name=HEATER2_4_2,output_id=21,output_type=outlet,type=output amp=1.2,state="AON",watt=138 1545716958000000000
> neptune_apex,device_id=4_3,host=ubuntu,hostname=apex,name=NUC_4_3,output_id=22,output_type=outlet,type=output amp=0.1,state="AON",watt=8 1545716958000000000
> neptune_apex,device_id=4_4,host=ubuntu,hostname=apex,name=CABFAN_4_4,output_id=23,output_type=outlet,type=output amp=0,state="AON",watt=1 1545716958000000000
> neptune_apex,device_id=4_5,host=ubuntu,hostname=apex,name=RHEAD_4_5,output_id=24,output_type=outlet,type=output amp=0,state="AON",watt=3 1545716958000000000
> neptune_apex,device_id=4_6,host=ubuntu,hostname=apex,name=FIRE_4_6,output_id=25,output_type=outlet,type=output amp=0,state="AON",watt=3 1545716958000000000
> neptune_apex,device_id=4_7,host=ubuntu,hostname=apex,name=LightGW_4_7,output_id=26,output_type=outlet,type=output amp=0,state="AON",watt=1 1545716958000000000
> neptune_apex,device_id=4_8,host=ubuntu,hostname=apex,name=GBSWITCH_4_8,output_id=27,output_type=outlet,type=output amp=0,state="AON",watt=1 1545716958000000000
> neptune_apex,device_id=4_9,host=ubuntu,hostname=apex,name=LinkA_4_9,output_id=28,output_type=unknown,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=4_10,host=ubuntu,hostname=apex,name=LinkB_4_10,output_id=29,output_type=unknown,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=5_1,host=ubuntu,hostname=apex,name=LinkA_5_1,output_id=30,output_type=unknown,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=Cntl_A1,host=ubuntu,hostname=apex,name=ATO_EMPTY,output_id=31,output_type=virtual,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=Cntl_A2,host=ubuntu,hostname=apex,name=LEAK,output_id=32,output_type=virtual,type=output state="AOF" 1545716958000000000
> neptune_apex,device_id=Cntl_A3,host=ubuntu,hostname=apex,name=SKMR_NOPWR,output_id=33,output_type=virtual,type=output state="AOF" 1545716958000000000
> neptune_apex,host=ubuntu,hostname=apex,name=Tmp,probe_type=Temp,type=probe value=78 1545716958000000000
> neptune_apex,host=ubuntu,hostname=apex,name=pH,probe_type=pH,type=probe value=7.98 1545716958000000000
> neptune_apex,host=ubuntu,hostname=apex,name=ORP,probe_type=ORP,type=probe value=193 1545716958000000000
> neptune_apex,host=ubuntu,hostname=apex,name=Salt,probe_type=Cond,type=probe value=33.8 1545716958000000000
> neptune_apex,host=ubuntu,hostname=apex,name=Volt_2,type=probe value=114 1545716958000000000
> neptune_apex,host=ubuntu,hostname=apex,name=Volt_4,type=probe value=115 1545716958000000000

```

### Contributing

This plugin is used for mission-critical aquatic life support. A bug could very well result in the death of animals.
Neptune does not publish a schema file and as such, we have made this plugin very strict on input with no provisions for 
automatically adding fields. We are also careful to not add default values when none are presented to prevent automation
errors.

When writing unit tests, use actual Apex output to run tests. It's acceptable to abridge the number of repeated fields 
but never inner fields or parameters. 