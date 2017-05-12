# sensors Input Plugin

Collect [lm-sensors](https://en.wikipedia.org/wiki/Lm_sensors) metrics - requires the lm-sensors
package installed.

This plugin collects sensor metrics with the `sensors` executable from the lm-sensor package.

### Configuration:
```
# Monitor sensors, requires lm-sensors package
[[inputs.sensors]]
```

### Measurements & Fields:
- All measurements have the following fields:
	- reading - This is the value of the sensor
	- unit - The unit of the sensor's value, C for temperature, V for voltage, RPM for fan speed

Fields are also generated for the additional information for each sensor (high, min, max, historical values).

### Tags:

- All measurements have the following tags:
    - chip
    - feature

### Sample Queries:

You can use the following query to get all temperature readings:

```
SELECT reading, chip, feature FROM sensors WHERE unit = 'C' GROUP BY chip, feature
```

### Example Output:

```
$ telegraf -config telegraf.conf -input-filter sensors -test
* Plugin: inputs.sensors, Collection 1
> sensors,chip=coretemp-isa-0000,feature=Core\ 0,host=HOST reading=50,unit="C",high=99 1494146884000000000
> sensors,feature=Core\ 1,host=HOST,chip=coretemp-isa-0001 reading=44,unit="C",high=99 1494146884000000000
> sensors,chip=coretemp-isa-0002,feature=Core\ 2,host=HOST reading=46,unit="C",high=99 1494146884000000000
> sensors,chip=coretemp-isa-0003,feature=Core\ 3,host=HOST reading=45,unit="C",high=99 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=VCore,host=HOST unit="V",min=0.6,max=1.49,reading=0.91 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=in1,host=HOST reading=11.88,unit="V",min=10.72,max=13.15 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=AVCC,host=HOST unit="V",min=2.96,max=3.63,reading=3.3 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=3VCC,host=HOST min=2.96,max=3.63,reading=3.3,unit="V" 1494146884000000000
> sensors,feature=in4,host=HOST,chip=w83627dhg-isa-0a10 max=1.65,reading=1.54,unit="V",min=1.35 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=in5,host=HOST min=1.13,max=1.38,reading=1.26,unit="V" 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=in6,host=HOST reading=4.66,unit="V",min=4.53,max=0 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=VSB,host=HOST unit="V",min=2.96,max=3.63,reading=3.3 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=VBAT,host=HOST min=2.96,max=3.63,reading=3.2,unit="V" 1494146884000000000
> sensors,feature=Case\ Fan,host=HOST,chip=w83627dhg-isa-0a10 reading=0,unit="RPM",min=10546,div=128 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=CPU\ Fan,host=HOST reading=5192,unit="RPM",min=5720,div=2 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=Aux\ Fan,host=HOST reading=0,unit="RPM",min=10546,div=128 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=fan4,host=HOST unit="RPM",min=10546,div=128,reading=0 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=fan5,host=HOST reading=0,unit="RPM",min=10546,div=128 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=Sys\ Temp,host=HOST reading=47,unit="C",high=60,hyst=55 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=CPU\ Temp,host=HOST reading=46.5,unit="C",high=95,hyst=92 1494146884000000000
> sensors,host=HOST,chip=w83627dhg-isa-0a10,feature=AUX\ Temp reading=46,unit="C",high=80,hyst=75 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=vid,host=HOST reading=1.3,unit="V" 1494146884000000000
> sensors,chip=w83627dhg-isa-0a10,feature=vid,host=HOST reading=1.3,unit="V" 1494146884000000000
```
