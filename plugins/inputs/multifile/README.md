# Multifile Input Plugin

### Description
The multifile input plugin allows telegraf to gather data from multiple files into a single point, creating one field or tag per file.

### Configuration
```
[[inputs.multifile]]
  ## Base directory where telegraf will look for files.
  ## Omit this option to use absolute paths.
  base_dir = "/sys/bus/i2c/devices/1-0076/iio:device0"

  ## If true, Telegraf discard all data when a single file can't be read.
  ## Else, Telegraf omits the field generated from this file.
  # fail_early = true

  ## Files to parse each interval.
  [[inputs.multifile.file]]
    file = "in_pressure_input"
    dest = "pressure"
    conversion = "float"
  [[inputs.multifile.file]]
    file = "in_temp_input"
    dest = "temperature"
    conversion = "float(3)"
  [[inputs.multifile.file]]
    file = "in_humidityrelative_input"
    dest = "humidityrelative"
    conversion = "float(3)"
```
* `file.file`:
Path of the file to be parsed
* `file.dest`:
Name of the field/tag created, defaults to `$(basename file)`
* `file.conversion`:
Data format used to parse the file contents
	* `float(X)`: Converts the input value into a float and divides by the Xth power of 10. Efficively just moves the decimal left X places. For example a value of `123` with `float(2)` will result in `1.23`.
	* `float`: Converts the value into a float with no adjustment. Same as `float(0)`.
	* `int`: Convertes the value into an integer.
	* `string`, `""`: No conversion
	* `bool`: Convertes the value into a boolean
	* `tag`: File content is used as a tag

### Example Output
This example shows a BME280 connected to a Raspberry Pi, using the sample config.
```
multifile pressure=101.343285156,temperature=20.4,humidityrelative=48.9 1547202076000000000
```

To reproduce this, connect a BMP280 to the board's GPIO pins and register the BME280 device driver
```
cd /sys/bus/i2c/devices/i2c-1
echo bme280 0x76 > new_device
```

The kernel driver provides the following files in `/sys/bus/i2c/devices/1-0076/iio:device0`:
* `in_humidityrelative_input`: `48900`
* `in_pressure_input`: `101.343285156`
* `in_temp_input`: `20400`
