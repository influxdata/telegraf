# DHT22 Input Plugin

Collect temperature and humidity from DHT22 device (usually on a raspberry pi)

example udev rules for the `gpio` group to have read/write access to `/sys/class/gpio#`:
*/etc/udev/rules.d/99-gpio.rules*

```
SUBSYSTEM=="gpio", ACTION=="add", PROGRAM="/bin/sh -c 'chown -R root:gpio /sys%p && chmod -R 770 /sys%p/*'"
SUBSYSTEM=="gpio", ACTION=="add", PROGRAM="/bin/sh -c 'chown -R root:gpio /sys/class/gpio && chmod -R 770 /sys/class/gpio'"
SUBSYSTEM=="gpio", ACTION=="add", PROGRAM="/bin/sh -c 'chown -R root:gpio /sys/devices/platform/soc/*.gpio && chmod -R 770 /sys/devices/platform/soc/*.gpio'"
```

### Configuration:
```
# Monitor sensors, requires lm-sensors package
[[inputs.dht22]]
  ## Set the GPIO pin
  pin = 14
  ## use boostPerfFlag
  boost = false
  ## how many times to retry for a good reading, should be over around 7.
  retry = 10
```

### Measurements & Fields:
Only 3 fields:

- temperature (float, degrees celsius)
- humidity (float, percentage)
- retries (integer, count)

### Example Output:

#### Default
```
$ telegraf --config telegraf.conf --input-filter hdt22 --test
* Plugin: dht22, Collection 1
> dht22,temperature=22.5,humidity=75.2,retries=3 1466751326000000000
```
