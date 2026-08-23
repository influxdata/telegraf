# Fritzbox Smarthome Input Plugin

This plugin gathers status information from Smarthome capable [FRITZ!][fritz]
routers using the device's [AVM Home Automation][aha] interface.

⭐ Telegraf v1.xx.x
🏷️ network, iot
💻 all

[fritz]: https://fritz.com/
[tr064]: https://fritz.com/en/pages/interfaces

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
    - `source` - The name of the device (this metric has been queried from)
    - `manufacturer` - The manufacturer of the smarthome device
    - `product_category` - The category of the smarthome device (e.g. sensor, button)
    - `power_source` - The power source of the smarthome device (internal, external, battery)
  - fields
    - `name`             (string) - Device's name (as given by user).
    - `product_name`     (string) - Device's product name.
    - `connected`        (bool)   - Device's connected status.
    - `battery_value`    (string) - Device's battery level (0-100).
    - `battery_low`      (bool)   - Device's battery low state.
    - `update_available` (bool)   - Device's software update available state.

## Example Output

```text
TODO
```
