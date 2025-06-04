# RAPL Input Plugin

This plugin reports the system energy consumption via the Intel RAPL interface.

⭐ Telegraf v1.35.0
🏷️ server,system
💻 linux

## Prerequisites

This plugin requires an Intel or AMD processor supporting the Intel RAPL
interface, and the associated Linux kernel modules. The system meets the
requirements if `/sys/devices/virtual/powercap` exists.

This plugin requires either that you run Telegraf as `root` (not recommended)
or that you grant unprivileged users read permissions to the RAPL energy
counters (recommended). The following instructions configure `udev` to grant
the necessary permissions at startup time.

Copy `rapl_init.sh` to `/usr/local/bin/rapl_init.sh`.

Create `/etc/udev/rules.d/99-rapl.rules` with the following contents.

```text
SUBSYSTEM=="powercap", RUN+="/usr/local/bin/rapl_init.sh"
```

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml
[[inputs.rapl]]
```

## Metrics

- rapl
  - tags:
    - name - the human-readable name of the power zone (e.g.: `package-0`)
    - power_zone - the system name of the power zone (e.g.: `intel-rapl:0`)
  - fields:
    - energy_joules - the energy counter

Note: The plugin reports the values of the raw RAPL energy counters in Joules.
To convert energy (Joules) into power (Watt), compute the derivative of the
energy with respect to time. A practical approach to estimate this derivative
is to calculate the energy difference between two consecutive metrics and
divide it by the time interval between them.

## Example Output

```text
rapl,name=package-0,power_zone=intel-rapl:0 energy_joules=55485.71523
rapl,name=core,power_zone=intel-rapl:0:0 energy_joules=564.946583
```
