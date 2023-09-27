# LM Sensors Input Plugin

Collect [lm-sensors](https://en.wikipedia.org/wiki/Lm_sensors) metrics -
requires the lm-sensors package installed.

This plugin collects sensor metrics with the `sensors` executable from the
lm-sensor package.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

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

### Tags

- All measurements have the following tags:
  - chip
  - feature

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
