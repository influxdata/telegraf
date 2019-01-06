# Device Input Plugin

#### Description
The device input plugin allows telegraf to read data from various hardware devices, which are not supported by the `lm_sensors` utility, such as the BMP280.
It is specifically intended for reading attached sensors on platforms such as the Raspberry Pi, but can be extended to read metrics from every device whose driver exposes values as individual files.

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

[[inputs.device]]
	type = "bme280"
	devices = ["/sys/bus/i2c/devices/1-0076/iio:device0"]
```
additional sensors of the same type can be added by adding their path to the devices list.
Additional sensors of a different type can be added by creating a new `[[inputs.device]]` field.

#### Internals / Adding new Sensors
The device input plugin reads a predefined list of files from every path specified in the `devices` list and applies different parsing and scaling operations for each file.
New device types can be added to the map inside the init function.
Each `DeviceField` specifies a file to be read, the field name for the resulting metric, the data format (int, float, string and bool supported) as well as a scaling factor, which is multiplied to the value if the data format is float.
