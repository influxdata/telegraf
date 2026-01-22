# Bond Input Plugin

This plugin collects metrics for both the network bond interface as well as its
slave interfaces using `/proc/net/bonding/*` files.

⭐ Telegraf v1.5.0
🏷️ system
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

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
  - tags:
    - `bond`: name of the bond
  - fields:
    - `active_slave`: currently active slave interface for active-backup mode
    - `status`: status of the interface (0: down , 1: up)

- bond_slave
  - tags:
    - `bond`: name of the bond
    - `interface`: name of the network interface
  - fields:
    - `failures`: amount of failures for bond's slave interface
    - `status`: status of the interface (0: down , 1: up)
    - `count`: number of slaves attached to bond
    - `actor_churned (for LACP bonds)`: count for local end of LACP bond flapped
    - `partner_churned (for LACP bonds)`: count for remote end of LACP bond flapped
    - `total_churned (for LACP bonds)`: full count of all churn events

- bond_sys
  - tags:
    - `bond`: name of the bond
    - `mode`: name of the bonding mode
  - fields:
    - `slave_count`: number of slaves
    - `ad_port_count`: number of ports

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
