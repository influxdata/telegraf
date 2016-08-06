# Hddtemp Input Plugin

This plugin reads data from hddtemp daemon

## Requirements

Hddtemp should be installed and its daemon running

## Configuration

```
[[inputs.hddtemp]]
## By default, telegraf gathers temps data from all disks detected by the
## hddtemp.
##
## Only collect temps from the selected disks.
##
## A * as the device name will return the temperature values of all disks.
##
# address = "127.0.0.1:7634"
# devices = ["sda", "*"]
```
