# Fritzbox Smarthome Input Plugin

This plugin gathers status information from Smarthome capable [FRITZ!][fritz]
routers using the device's [AVM Home Automation][aha] interface.

⭐ Telegraf v1.40.0
🏷️ network, iot
💻 all

[fritz]: https://fritz.com/
[aha]: https://fritz.com/en/pages/interfaces

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather fritzbox smarthome status
[[inputs.fritzbox_smarthome]]
  ## URLs of the devices to query including login credentials
  urls = [ "http://user:password@fritz.box/" ]

  ## The http timeout to use.
  # timeout = "10s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  # tls_key_pwd = "secret"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Metrics

By default field names are directly derived from the corresponding [interface
specification][aha].

- `fritzbox_smarthome_device`
  - tags
    - `source`           - The hostname of the router
    - `manufacturer`     - The manufacturer of the smarthome device
    - `product_category` - The category (e.g. sensor, button)
    - `power_source`     - The power source (internal, external, battery)
  - fields
    - `name`             (string) - Device's name (as given by user).
    - `product_name`     (string) - Device's product name.
    - `connected`        (int)    - Device's connected status (0|1).
    - `battery_value`    (int)    - Device's battery level (0-100).
    - `battery_low`      (int)    - Device's battery low state (0|1).
    - `update_available` (int)    - Device's software update available state (0|1).
- `fritzbox_smarthome_multimeter`
  - tags
    - `source` - The name of the router (this metric has been queried from)
    - `group`  - The group this unit is assigned to (&lt;none&gt; if unassigned)
    - `type`   - The type of the unit (as reported by the API; e.g. dimmableLight)
  - fields
    - `name`    (string) - The name of the unit or group.
    - `energy`  (int)    - Overall energy consumption (Wh).
    - `power`   (int)    - Latest power consumption (mW).
    - `voltage` (int)    - Latest voltage (mV).
    - `current` (int)    - Latest current (mA).
- `fritzbox_smarthome_level_control`
  - tags
    - `source` - The name of the router (this metric has been queried from)
    - `group`  - The group this unit is assigned to (&lt;none&gt; if unassigned)
    - `type`   - The type of the unit (as reported by the API; e.g. dimmableLight)
  - fields
    - `name`  (string) - The name of the unit or group.
    - `level` (int)    - The current light level (0-100).
- `fritzbox_smarthome_on_off`
  - tags
    - `source` - The name of the router (this metric has been queried from)
    - `group`  - The group this unit is assigned to (&lt;none&gt; if unassigned)
    - `type`   - The type of the unit (as reported by the API; e.g. dimmableLight)
  - fields
    - `name`  (string) - The name of the unit or group.
    - `active`(int)    - The current active state (0|1).

## Example Output

```text
fritzbox_smarthome_device,manufacturer=AVM,power_source=battery,product_category=sensor,source=127.0.0.1 connected=1i,battery_value=100i,battery_low=0i,update_available=0i,name="Name#1",product_name="FRITZ!Smart Energy 250" 1787604406292123000

fritzbox_smarthome_multimeter,group=<none>,source=127.0.0.1,type=avmMeter energy=325528i,power=289130i,voltage=0i,name="Name#1",current=0i 1787604406292152000

fritzbox_smarthome_level_control,group=Name#3,source=127.0.0.1,type=dimmableLight name="Name#2",level=100i 1787604406292154000

fritzbox_smarthome_on_off,group=Name#3,source=127.0.0.1,type=dimmableLight name="Name#2",active=0i 1787604406292156000
```
