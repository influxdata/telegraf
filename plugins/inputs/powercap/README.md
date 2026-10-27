# Powercap Input Plugin

This plugin reads the energy counters of the power zones in the Linux
[power capping framework][powercap] and reports the average power since the
previous reading. On AMD processors these zones are the Running Average Power
Limit (RAPL) domains.

⭐ Telegraf v1.41.0
🏷️ hardware, system
💻 linux

[powercap]: https://docs.kernel.org/power/powercap/powercap.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read energy consumption from the Linux powercap framework
[[inputs.powercap]]
  ## No configuration options
```

The plugin reads the zones in `/sys/class/powercap`. Set the `HOST_SYS`
environment variable to read them from another location, for example when
Telegraf runs in a container with the host's `/sys` mounted at `/hostfs/sys`.

### Permissions

Since Linux 5.10, and in stable kernels since 5.4.77, only `root` can read the
energy counters by default because of the [PLATYPUS][platypus] side-channel
attack. Without read access the plugin reports a `permission denied` error for
each zone.

To let Telegraf read the counters without running as `root`, give the
`telegraf` group read access with a udev rule, for example in
`/etc/udev/rules.d/99-powercap.rules`:

```text
SUBSYSTEM=="powercap", TEST=="energy_uj", RUN+="/bin/chgrp telegraf /sys%p/energy_uj", RUN+="/bin/chmod g+r /sys%p/energy_uj"
```

Any process in the `telegraf` group can then read the counters. Decide if this
is acceptable for your system.

[platypus]: https://platypusattack.com

## Metrics

- powercap
  - tags:
    - zone (string) - The kernel's identifier of the power zone, for example
      `intel-rapl:0` or `intel-rapl:0:0`.
    - name (string) - The name of the power zone, for example `package-0`,
      `core` or `dram`.
  - fields:
    - energy_uj (uint64, microjoules) - The energy counter of the zone.
    - power_watts (float, watts) - The average power since the previous
      reading. The plugin reports it from the second reading onwards.

> [!NOTE]
> The energy counter overflows back to 0 at the maximum the kernel reports in
> `max_energy_range_uj`. On a CPU package that uses 280 W and has a range of
> 65,532 J, the counter restarts about every 4 minutes. Keep the collection
> `interval` well below the restart time of your system.

## Example Output

```text
powercap,name=package-0,zone=intel-rapl:0 energy_uj=25052989915i,power_watts=112.76534514584188 1793075282000000000
powercap,name=core,zone=intel-rapl:0:0 energy_uj=14860540833i,power_watts=10.395216507698843 1793075282000000000
```
