# Hddtemp Input Plugin

This plugin reads data from hddtemp daemon

## Requirements

Hddtemp should be installed and its daemon running

## Configuration

```toml
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

## Measurements

- hddtemp
  - temperature

Tags:
- device
- model
- unit
- status



## Example output

```
> hddtemp,unit=C,status=,host=server1,device=sdb,model=WDC\ WD740GD-00FLA1 temperature=43i 1481655647000000000
> hddtemp,device=sdc,model=SAMSUNG\ HD103UI,unit=C,status=,host=server1 temperature=38i 148165564700000000
> hddtemp,device=sdd,model=SAMSUNG\ HD103UI,unit=C,status=,host=server1 temperature=36i 1481655647000000000
```
