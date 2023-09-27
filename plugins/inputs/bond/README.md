# Bond Input Plugin

The Bond input plugin collects network bond interface status for both the
network bond interface as well as slave interfaces.
The plugin collects these metrics from `/proc/net/bonding/*` files.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Collect bond interface status, slaves statuses and failures count
[[inputs.bond]]
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  # host_proc = "/proc"

  ## Sets 'sys' directory path
  ## If not specified, then default is /sys
  # host_sys = "/sys"

  ## By default, telegraf gather stats for all bond interfaces
  ## Setting interfaces will restrict the stats to the specified
  ## bond interfaces.
  # bond_interfaces = ["bond0"]

  ## Tries to collect additional bond details from /sys/class/net/{bond}
  ## currently only useful for LACP (mode 4) bonds
  # collect_sys_details = false
```

## Metrics

- bond
  - active_slave (for active-backup mode)
  - status

- bond_slave
  - failures
  - status
  - count
  - actor_churned (for LACP bonds)
  - partner_churned (for LACP bonds)
  - total_churned (for LACP bonds)

- bond_sys
  - slave_count
  - ad_port_count

## Description

- active_slave
  - Currently active slave interface for active-backup mode.
- status
  - Status of bond interface or bonds's slave interface (down = 0, up = 1).
- failures
  - Amount of failures for bond's slave interface.
- count
  - Number of slaves attached to bond
- actor_churned
  - number of times local end of LACP bond flapped
- partner_churned
  - number of times remote end of LACP bond flapped
- total_churned
  - full count of all churn events

## Tags

- bond
  - bond

- bond_slave
  - bond
  - interface

- bond_sys
  - bond
  - mode

## Example Output

Configuration:

```toml
[[inputs.bond]]
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  host_proc = "/proc"

  ## By default, telegraf gather stats for all bond interfaces
  ## Setting interfaces will restrict the stats to the specified
  ## bond interfaces.
  bond_interfaces = ["bond0", "bond1"]
```

Run:

```bash
telegraf --config telegraf.conf --input-filter bond --test
```

Output:

```text
bond,bond=bond1,host=local active_slave="eth0",status=1i 1509704525000000000
bond_slave,bond=bond1,interface=eth0,host=local status=1i,failures=0i 1509704525000000000
bond_slave,host=local,bond=bond1,interface=eth1 status=1i,failures=0i 1509704525000000000
bond_slave,host=local,bond=bond1 count=2i 1509704525000000000
bond,bond=bond0,host=isvetlov-mac.local status=1i 1509704525000000000
bond_slave,bond=bond0,interface=eth1,host=local status=1i,failures=0i 1509704525000000000
bond_slave,bond=bond0,interface=eth2,host=local status=1i,failures=0i 1509704525000000000
bond_slave,bond=bond0,host=local count=2i 1509704525000000000
```
