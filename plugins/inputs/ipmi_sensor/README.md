# IPMI Sensor Input Plugin

Get bare metal metrics using the command line utility
[`ipmitool`](https://github.com/ipmitool/ipmitool).

If no servers are specified, the plugin will query the local machine sensor
stats via the following command:

```sh
ipmitool sdr
```

or with the version 2 schema:

```sh
ipmitool sdr elist
```

When one or more servers are specified, the plugin will use the following
command to collect remote host sensor stats:

```sh
ipmitool -I lan -H SERVER -U USERID -P PASSW0RD sdr
```

Any of the following parameters will be added to the aformentioned query if
they're configured:

```sh
-y hex_key -L privilege
```

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from the bare metal servers via IPMI
[[inputs.ipmi_sensor]]
  ## optionally specify the path to the ipmitool executable
  # path = "/usr/bin/ipmitool"
  ##
  ## Setting 'use_sudo' to true will make use of sudo to run ipmitool.
  ## Sudo must be configured to allow the telegraf user to run ipmitool
  ## without a password.
  # use_sudo = false
  ##
  ## optionally force session privilege level. Can be CALLBACK, USER, OPERATOR, ADMINISTRATOR
  # privilege = "ADMINISTRATOR"
  ##
  ## optionally specify one or more servers via a url matching
  ##  [username[:password]@][protocol[(address)]]
  ##  e.g.
  ##    root:passwd@lan(127.0.0.1)
  ##
  ## if no servers are specified, local machine sensor stats will be queried
  ##
  # servers = ["USERID:PASSW0RD@lan(192.168.1.1)"]

  ## Recommended: use metric 'interval' that is a multiple of 'timeout' to avoid
  ## gaps or overlap in pulled data
  interval = "30s"

  ## Timeout for the ipmitool command to complete. Default is 20 seconds.
  timeout = "20s"

  ## Schema Version: (Optional, defaults to version 1)
  metric_version = 2

  ## Optionally provide the hex key for the IMPI connection.
  # hex_key = ""

  ## If ipmitool should use a cache
  ## for me ipmitool runs about 2 to 10 times faster with cache enabled on HP G10 servers (when using ubuntu20.04)
  ## the cache file may not work well for you if some sensors come up late
  # use_cache = false

  ## Path to the ipmitools cache file (defaults to OS temp dir)
  ## The provided path must exist and must be writable
  # cache_path = ""
```

## Metrics

Version 1 schema:

- ipmi_sensor:
  - tags:
    - name
    - unit
    - host
    - server (only when retrieving stats from remote servers)
  - fields:
    - status (int, 1=ok status_code/0=anything else)
    - value (float)

Version 2 schema:

- ipmi_sensor:
  - tags:
    - name
    - entity_id (can help uniquify duplicate names)
    - status_code (two letter code from IPMI documentation)
    - status_desc (extended status description field)
    - unit (only on analog values)
    - host
    - server (only when retrieving stats from remote)
  - fields:
    - value (float)

### Permissions

When gathering from the local system, Telegraf will need permission to the
ipmi device node.  When using udev you can create the device node giving
`rw` permissions to the `telegraf` user by adding the following rule to
`/etc/udev/rules.d/52-telegraf-ipmi.rules`:

```sh
KERNEL=="ipmi*", MODE="660", GROUP="telegraf"
```

Alternatively, it is possible to use sudo. You will need the following in your
telegraf config:

```toml
[[inputs.ipmi_sensor]]
  use_sudo = true
```

You will also need to update your sudoers file:

```bash
$ visudo
# Add the following line:
Cmnd_Alias IPMITOOL = /usr/bin/ipmitool *
telegraf  ALL=(root) NOPASSWD: IPMITOOL
Defaults!IPMITOOL !logfile, !syslog, !pam_session
```

## Example Output

### Version 1 Schema

When retrieving stats from a remote server:

```text
ipmi_sensor,server=10.20.2.203,name=uid_light value=0,status=1i 1517125513000000000
ipmi_sensor,server=10.20.2.203,name=sys._health_led status=1i,value=0 1517125513000000000
ipmi_sensor,server=10.20.2.203,name=power_supply_1,unit=watts status=1i,value=110 1517125513000000000
ipmi_sensor,server=10.20.2.203,name=power_supply_2,unit=watts status=1i,value=120 1517125513000000000
ipmi_sensor,server=10.20.2.203,name=power_supplies value=0,status=1i 1517125513000000000
ipmi_sensor,server=10.20.2.203,name=fan_1,unit=percent status=1i,value=43.12 1517125513000000000
```

When retrieving stats from the local machine (no server specified):

```text
ipmi_sensor,name=uid_light value=0,status=1i 1517125513000000000
ipmi_sensor,name=sys._health_led status=1i,value=0 1517125513000000000
ipmi_sensor,name=power_supply_1,unit=watts status=1i,value=110 1517125513000000000
ipmi_sensor,name=power_supply_2,unit=watts status=1i,value=120 1517125513000000000
ipmi_sensor,name=power_supplies value=0,status=1i 1517125513000000000
ipmi_sensor,name=fan_1,unit=percent status=1i,value=43.12 1517125513000000000
```

#### Version 2 Schema

When retrieving stats from the local machine (no server specified):

```text
ipmi_sensor,name=uid_light,entity_id=23.1,status_code=ok,status_desc=ok value=0 1517125474000000000
ipmi_sensor,name=sys._health_led,entity_id=23.2,status_code=ok,status_desc=ok value=0 1517125474000000000
ipmi_sensor,entity_id=10.1,name=power_supply_1,status_code=ok,status_desc=presence_detected,unit=watts value=110 1517125474000000000
ipmi_sensor,name=power_supply_2,entity_id=10.2,status_code=ok,unit=watts,status_desc=presence_detected value=125 1517125474000000000
ipmi_sensor,name=power_supplies,entity_id=10.3,status_code=ok,status_desc=fully_redundant value=0 1517125474000000000
ipmi_sensor,entity_id=7.1,name=fan_1,status_code=ok,status_desc=transition_to_running,unit=percent value=43.12 1517125474000000000
```
