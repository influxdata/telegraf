# HDDtemp Input Plugin

This plugin reads data from a [hddtemp][hddtemp] daemon.

> [!IMPORTANT]
> This plugin requires `hddtemp` to be installed and running as a daemon.

As the upstream project is not activly maintained anymore and various
distributions (e.g. Debian Bookwork and later) don't ship packages for `hddtemp`
anymore, the binary might not be available (e.g. in Ubuntu 22.04 or later).

> [!TIP]
> As an alternative consider using the [smartctl][smartctl] relying on
> SMART information or [sensors][sensors] plugins to retrieve temperature data
> of your hard-drive.

⭐ Telegraf v1.0.0
🏷️ hardware, system
💻 all

[hddtemp]: https://savannah.nongnu.org/projects/hddtemp/
[smartctl]: /plugins/inputs/smartctl/README.md
[sensors]: /plugins/inputs/sensors/README.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Monitor disks' temperatures using hddtemp
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

## Metrics

- hddtemp
  - tags:
    - device
    - model
    - unit
    - status
    - source
  - fields:
    - temperature

## Example Output

```text
hddtemp,source=server1,unit=C,status=,device=sdb,model=WDC\ WD740GD-00FLA1 temperature=43i 1481655647000000000
hddtemp,device=sdc,model=SAMSUNG\ HD103UI,unit=C,source=server1,status= temperature=38i 148165564700000000
hddtemp,device=sdd,model=SAMSUNG\ HD103UI,unit=C,source=server1,status= temperature=36i 1481655647000000000
```
