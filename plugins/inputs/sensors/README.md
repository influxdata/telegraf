# LM Sensors Input Plugin

This plugin collects metrics from hardware sensors using
[lm-sensors][lmsensors].

> [!NOTE]
> This plugin requires the lm-sensors package to be installed on the system
> and `sensors` to be executable from Telegraf.

⭐ Telegraf v0.10.1
🏷️ hardware, system
💻 linux

[lmsensors]: https://en.wikipedia.org/wiki/Lm_sensors

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Monitor sensors, requires lm-sensors package
# This plugin ONLY supports Linux
[[inputs.sensors]]
  ## Remove numbers from field names.
  ## If true, a field name like 'temp1_input' will be changed to 'temp_input'.
  # remove_numbers = true

  ## Timeout is the maximum amount of time that the sensors command can run.
  # timeout = "5s"
```

## Metrics

Fields are created dynamically depending on the sensors. All fields are float.

- sensors:
  - tags:
    - chip
    - feature
  - fields:
    - depending on the available sensor information (float)

## Example Output

### Default

```text
sensors,chip=power_meter-acpi-0,feature=power1 power_average=0,power_average_interval=300 1466751326000000000
sensors,chip=k10temp-pci-00c3,feature=temp1 temp_crit=70,temp_crit_hyst=65,temp_input=29,temp_max=70 1466751326000000000
sensors,chip=k10temp-pci-00cb,feature=temp1 temp_input=29,temp_max=70 1466751326000000000
sensors,chip=k10temp-pci-00d3,feature=temp1 temp_input=27.5,temp_max=70 1466751326000000000
sensors,chip=k10temp-pci-00db,feature=temp1 temp_crit=70,temp_crit_hyst=65,temp_input=29.5,temp_max=70 1466751326000000000
```

### With remove_numbers=false

```text
sensors,chip=power_meter-acpi-0,feature=power1 power1_average=0,power1_average_interval=300 1466753424000000000
sensors,chip=k10temp-pci-00c3,feature=temp1 temp1_crit=70,temp1_crit_hyst=65,temp1_input=29.125,temp1_max=70 1466753424000000000
sensors,chip=k10temp-pci-00cb,feature=temp1 temp1_input=29,temp1_max=70 1466753424000000000
sensors,chip=k10temp-pci-00d3,feature=temp1 temp1_input=29.5,temp1_max=70 1466753424000000000
sensors,chip=k10temp-pci-00db,feature=temp1 temp1_crit=70,temp1_crit_hyst=65,temp1_input=30,temp1_max=70 1466753424000000000
```
