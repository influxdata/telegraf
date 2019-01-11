# Multifile Input Plugin

#### Description
The multifile input plugin allows telegraf to gather data from multiple files into a single point, creating one field or tag per file.

#### Configuration
* `base_dir`:
Base directory for all files. If empty, all file paths are seen as absolute.
* `tags`:
Table of additional tags.
* `file.file`:
Filename, relative to `base_dir`
* `file.dest`:
Name of the field/tag created, defaults to `$(basename file)`
* `file.conversion`:
Data format used to parse the file contents

    - `float(X)`: Converts the input value into a float and divides by the Xth power of 10. Efficively just moves the decimal left X places. For example a value of `123` with `float(2)` will result in `1.23`.
    - `float`: Converts the value into a float with no adjustment. Same as `float(0)`.
    - `int`: Convertes the value into an integer.
    - `string`, `""`: No conversion
    - `bool`: Convertes the value into a boolean
    - `tag`: File content is used as a tag

#### Example for using a BMP280 with the Raspberry Pi 3B
Connect a BMP280 to the board's GPIO pins and register the BME280 device driver
```
sudo su
cd /sys/bus/i2c/devices/i2c-1
echo bme280 0x76 > new_device
```
or using `/boot/config.txt`
```
dtparam=i2c_arm=on
dtoverlay=i2c-sensor,bme280
```
and start telegraf with the following configuration
`telegraf.conf`
```
[agent]
	interval = "1s"
	flush_interval = "1s"

[[outputs.file]]
	files = ["stdout"]
	data_format = "influx"

[[inputs.multifile]]
  name_override = "sensor"
  base_dir = "/sys/bus/i2c/devices/1-0076/iio:device0"

  [inputs.multifile.tags]
    location = "server_room"

  [[inputs.multifile.file]]
    file = "name"
    dest = "type"
    conversion = "tag"

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

this will read the following files from `/sys/bus/i2c/devices/1-0076/iio:device0`:
* `in_humidityrelative_input`: `48900`
* `in_pressure_input`: `101.343285156`
* `in_temp_input`: `20400`
* `name`: `bmp280`

and output:
```
sensor,host=raspberry,location=server_room,type=bme280 pressure=101.343285156,temperature=20.4,humidityrelative=48.9 1547202076000000000
```
