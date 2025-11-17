# Turbostat Input Plugin

This service plugin monitors system performance using the [turbostat][turbostat]
command.

> [!IMPORTANT]
> This plugin requires the `turbostat` executable to be installed on the system.

⭐ Telegraf v1.36.0
🏷️ hardware,system
💻 linux

[turbostat]: https://github.com/torvalds/linux/tree/master/tools/power/x86/turbostat

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listen and wait for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
## Gather CPU metrics using Turbostat
[[inputs.turbostat]]
  ## Path to the Turbostat exectuable if not in the PATH
  # path = "/usr/bin/turbostat"

  ## Turbostat measurement interval
  # interval = "10s"

  ## Use sudo to run the Turbostat executable
  # use_sudo = false
```

Allow the Telegraf user to run `turbostat`. Assuming the Telegraf user is
`telegraf` and `turbostat` is installed in `/usr/bin`, add the following line
to `/etc/sudoers.d/telegraf`.

```text
telegraf ALL=(root) NOPASSWD: /usr/bin/turbostat
```

Some Turbostat configuration options are particularly relevant for this plugin.

- `-i`, `--interval`: Overrides the default 5s collection interval. This plugin
is a service input plugin: it ignores the Telegraf global and plugin-specific
collection intervals. The interval should be configured in Turbostat instead.
- `-S`, `--summary`: Limits the output to a one-line system summary per
collection interval.
- `-s`, `--show`: Show only the specified columns. May be invoked multiple
times, or with a comma-separated list of column names.
- `-H`, `--hide`: Do not show the specified columns. May be invoked multiple
times, or with a comma-separated list of column names. Be careful not to hide
CPU, Core, Package, and Die, or the output may lose much of its meaning.

To discover the list of available columns, run `sudo turbostat --list`.

For further information, run `man turbostat`. If the man page is not installed,
download [turbostat.8][turbostat.8] and browse it with `man ./turbostat.8`.

[turbostat.8]: https://raw.githubusercontent.com/torvalds/linux/refs/heads/master/tools/power/x86/turbostat/turbostat.8

## Metrics

The exact metrics depend on the version of Turbostat, the system on which it
runs, and the command line parameters. For further information, browse the
Turbostat documentation.

## Example Output

On an AMD consumer system.

```text
turbostat,apic=-,core=-,cpu=-,x2apic=- average_frequency=41000000,busy_frequency=3431000000,busy_percent=1.2,c1=2836,c1_percent=0.33,c2=10698,c2_percent=5.55,c3=15564,c3_percent=93.03,core_power=1.12,ipc=0.98,irq=30422,package_power=24.55,poll=1412,poll_percent=0.04,tsc_frequency=3793000000,usec=966
turbostat,apic=0,core=0,cpu=0,x2apic=0 average_frequency=34000000,busy_frequency=3083000000,busy_percent=1.09,c1=102,c1_percent=0.57,c2=785,c2_percent=6.81,c3=1282,c3_percent=91.68,core_power=0.14,ipc=0.92,irq=2306,package_power=24.55,poll=63,poll_percent=0.04,tsc_frequency=3793000000,usec=54
turbostat,apic=1,core=0,cpu=8,x2apic=1 average_frequency=27000000,busy_frequency=3522000000,busy_percent=0.77,c1=293,c1_percent=0.37,c2=899,c2_percent=5.49,c3=668,c3_percent=93.47,ipc=1.11,irq=1781,poll=78,poll_percent=0.03,tsc_frequency=3793000000,usec=30
turbostat,apic=2,core=1,cpu=1,x2apic=2 average_frequency=91000000,busy_frequency=3665000000,busy_percent=2.49,c1=205,c1_percent=0.72,c2=1087,c2_percent=10.56,c3=1342,c3_percent=86.4,core_power=0.27,ipc=0.78,irq=2867,poll=165,poll_percent=0.08,tsc_frequency=3793000000,usec=69
turbostat,apic=3,core=1,cpu=9,x2apic=3 average_frequency=112000000,busy_frequency=3851000000,busy_percent=2.92,c1=51,c1_percent=0.24,c2=703,c2_percent=6.79,c3=1197,c3_percent=90.19,ipc=1.13,irq=2441,poll=64,poll_percent=0.05,tsc_frequency=3793000000,usec=76
turbostat,apic=4,core=2,cpu=2,x2apic=4 average_frequency=38000000,busy_frequency=3469000000,busy_percent=1.09,c1=105,c1_percent=0.4,c2=646,c2_percent=6.18,c3=1240,c3_percent=92.48,core_power=0.15,ipc=0.87,irq=2192,poll=89,poll_percent=0.06,tsc_frequency=3793000000,usec=69
turbostat,apic=5,core=2,cpu=10,x2apic=5 average_frequency=20000000,busy_frequency=3488000000,busy_percent=0.56,c1=107,c1_percent=0.39,c2=694,c2_percent=5.36,c3=908,c3_percent=93.8,ipc=0.97,irq=1794,poll=73,poll_percent=0.04,tsc_frequency=3793000000,usec=50
turbostat,apic=6,core=3,cpu=3,x2apic=6 average_frequency=105000000,busy_frequency=3546000000,busy_percent=2.96,c1=563,c1_percent=0.63,c2=908,c2_percent=8.5,c3=1130,c3_percent=88.05,core_power=0.17,ipc=1.29,irq=2545,poll=477,poll_percent=0.06,tsc_frequency=3793000000,usec=40
turbostat,apic=7,core=3,cpu=11,x2apic=7 average_frequency=21000000,busy_frequency=3250000000,busy_percent=0.65,c1=931,c1_percent=0.38,c2=542,c2_percent=3.99,c3=623,c3_percent=95.07,ipc=0.78,irq=1293,poll=46,poll_percent=0.01,tsc_frequency=3793000000,usec=35
turbostat,apic=8,core=4,cpu=4,x2apic=8 average_frequency=29000000,busy_frequency=3044000000,busy_percent=0.94,c1=57,c1_percent=0.07,c2=880,c2_percent=4.87,c3=894,c3_percent=94.24,core_power=0.1,ipc=0.85,irq=1948,poll=37,poll_percent=0.03,tsc_frequency=3793000000,usec=70
```

On an Intel server system.

```text
turbostat,core=-,cpu=-,die=-,package=- average_frequency=1000000,busy_frequency=1321000000,busy_percent=0.04,c1=1,c1_percent=0,c1e=1909,c1e_percent=100.01,c3=0,c3_percent=0,c6=0,c6_percent=0,core_temperature=48,core_throttle=0,cpu_percent_c1=99.96,cpu_percent_c3=0,cpu_percent_c6=0,ipc=1.01,irq=1971,package_percent=0,package_percent_pc2=0,package_percent_pc3=0,package_percent_pc6=0,package_power=87.38,package_temperature=53,poll=1,poll_percent=0,ram_percent=0,ram_power=27.8,smi=0,tsc_frequency=3001000000,uncore_frequency=2800000000
turbostat,core=0,cpu=0,die=0,package=0 average_frequency=1000000,busy_frequency=1257000000,busy_percent=0.07,c1=0,c1_percent=0,c1e=69,c1e_percent=99.93,c3=0,c3_percent=0,c6=0,c6_percent=0,core_temperature=46,core_throttle=0,cpu_percent_c1=99.93,cpu_percent_c3=0,cpu_percent_c6=0,ipc=0.82,irq=53,package_percent=0,package_percent_pc2=0,package_percent_pc3=0,package_percent_pc6=0,package_power=45.14,package_temperature=53,poll=0,poll_percent=0,ram_percent=0,ram_power=13.66,smi=0,tsc_frequency=3000000000,uncore_frequency=2800000000
turbostat,core=0,cpu=24,die=0,package=0 average_frequency=0,busy_frequency=1398000000,busy_percent=0.03,c1=0,c1_percent=0,c1e=41,c1e_percent=99.97,c3=0,c3_percent=0,c6=0,c6_percent=0,cpu_percent_c1=99.97,ipc=0.59,irq=41,poll=0,poll_percent=0,smi=0,tsc_frequency=3000000000
turbostat,core=1,cpu=2,die=0,package=0 average_frequency=0,busy_frequency=1200000000,busy_percent=0.01,c1=0,c1_percent=0,c1e=3,c1e_percent=99.99,c3=0,c3_percent=0,c6=0,c6_percent=0,core_temperature=48,core_throttle=0,cpu_percent_c1=99.99,cpu_percent_c3=0,cpu_percent_c6=0,ipc=0.72,irq=1,poll=0,poll_percent=0,smi=0,tsc_frequency=3000000000
turbostat,core=1,cpu=26,die=0,package=0 average_frequency=0,busy_frequency=1200000000,busy_percent=0.04,c1=0,c1_percent=0,c1e=96,c1e_percent=99.97,c3=0,c3_percent=0,c6=0,c6_percent=0,cpu_percent_c1=99.96,ipc=0.88,irq=95,poll=0,poll_percent=0,smi=0,tsc_frequency=3000000000
turbostat,core=2,cpu=4,die=0,package=0 average_frequency=0,busy_frequency=1200000000,busy_percent=0.01,c1=0,c1_percent=0,c1e=3,c1e_percent=99.99,c3=0,c3_percent=0,c6=0,c6_percent=0,core_temperature=48,core_throttle=0,cpu_percent_c1=99.99,cpu_percent_c3=0,cpu_percent_c6=0,ipc=0.7,irq=2,poll=0,poll_percent=0,smi=0,tsc_frequency=3000000000
turbostat,core=2,cpu=28,die=0,package=0 average_frequency=0,busy_frequency=1200000000,busy_percent=0.04,c1=0,c1_percent=0,c1e=38,c1e_percent=99.96,c3=0,c3_percent=0,c6=0,c6_percent=0,cpu_percent_c1=99.96,ipc=0.97,irq=37,poll=0,poll_percent=0,smi=0,tsc_frequency=3000000000
turbostat,core=3,cpu=6,die=0,package=0 average_frequency=0,busy_frequency=1200000000,busy_percent=0.02,c1=0,c1_percent=0,c1e=22,c1e_percent=99.98,c3=0,c3_percent=0,c6=0,c6_percent=0,core_temperature=47,core_throttle=0,cpu_percent_c1=99.98,cpu_percent_c3=0,cpu_percent_c6=0,ipc=0.69,irq=20,poll=0,poll_percent=0,smi=0,tsc_frequency=3000000000
turbostat,core=3,cpu=30,die=0,package=0 average_frequency=1000000,busy_frequency=1200000000,busy_percent=0.05,c1=0,c1_percent=0,c1e=78,c1e_percent=99.96,c3=0,c3_percent=0,c6=0,c6_percent=0,cpu_percent_c1=99.95,ipc=0.71,irq=80,poll=0,poll_percent=0,smi=0,tsc_frequency=3000000000
turbostat,core=4,cpu=8,die=0,package=0 average_frequency=0,busy_frequency=1200000000,busy_percent=0.01,c1=0,c1_percent=0,c1e=3,c1e_percent=99.99,c3=0,c3_percent=0,c6=0,c6_percent=0,core_temperature=47,core_throttle=0,cpu_percent_c1=99.99,cpu_percent_c3=0,cpu_percent_c6=0,ipc=0.72,irq=1,poll=0,poll_percent=0,smi=0,tsc_frequency=3000000000
```
