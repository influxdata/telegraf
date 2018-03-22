# DHT22 Input Plugin

Collect temperature and humidity from DHT22 sensor (usually on a raspberry pi)


example udev rules for read/write access to `/sys/class/gpio#`:
*/etc/udev/rules.d/99-gpio.rules*

```
SUBSYSTEM=="gpio", ACTION=="add", PROGRAM="/bin/sh -c 'chmod -R 777 /sys%p/*'"
SUBSYSTEM=="gpio", ACTION=="add", PROGRAM="/bin/sh -c 'chmod -R 777 /sys/class/gpio'"
SUBSYSTEM=="gpio", ACTION=="add", PROGRAM="/bin/sh -c 'chmod -R 777 /sys/devices/platform/soc/*.gpio'"
```

### Configuration:
```
# Monitor sensors, requires lm-sensors package
[[inputs.dht22]]
  ## Set the GPIO pin
  pin = 14
  ## how many times to retry for a good reading.
  retry = 10
  ## Additionally calculate Vapor Pressure Deficit in kPa
  calcvpd = true
  ## divisor/multiplier for VPD (1000 to transform to Pa)
  vpdmultiplier = 1
```

### Measurements & Fields:
4 fields:

- temperature (float, degrees celsius)
- humidity (float, percentage)
- vpd (float, kPa)
- retries (integer, count)

### Example Output:

#### Default
```
$ telegraf --config telegraf.conf --input-filter dht22 --test
* Plugin: dht22, Collection 1
> dht22,temperature=22.5,humidity=75.2,retries=3 1466751326000000000
```
