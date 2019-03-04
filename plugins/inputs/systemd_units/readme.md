# Systemd Units Plugin

The systemd_units plugin gathers systemd unit status on Linux. It relies on
```systemctl list-units --type=service``` to collect data on service status.

The results are tagged with the unit name and provide enumerated fields for
loaded, active and running fields, indicating the unit health.

This plugin is related to the win_services module, which fulfills the same
purpose on windows.

In addition to services, this plugin can gather other unit types as well,
see ```systemctl list-units --type help``` for possible options.

### Configuration
```
[[inputs.systemd_units]]
  ## The default timeout of 1s for systemctl execution can be overridden here:
  # timeout = "1s"
  # unittype = "service"
```

### Metrics
- systemd_units:
  - tags:
    - name (string, unit name)
  - fields:
    - loaded (int, see below)
    - active (int, see below)
    - running (int, see below)

#### Loaded

| Value | Meaning     | Description                     |
| ----- | -------     | -----------                     |
| 0     | loaded      | unit is ~                       |
| 1     | stub        | unit is ~                       |
| 2     | not-found   | unit is ~                       |
| 3     | bad-setting | unit is ~                       |
| 4     | error       | unit is ~                       |
| 5     | merged      | unit is ~                       |
| 6     | masked      | unit is ~                       |
| -1    | err         | field not found in lookup table |

#### Active

| Value | Meaning   | Description                        |
| ----- | -------   | -----------                        |
| 0     | active       | unit is ~                       |
| 1     | reloading    | unit is ~                       |
| 2     | inactive     | unit is ~                       |
| 3     | failed       | unit is ~                       |
| 4     | activating   | unit is ~                       |
| 5     | deactivating | unit is ~                       |
| -1    | err          | field not found in lookup table |

#### Running

| Value  | Meaning               | Description                         |
| -----  | -------               | -----------                         |
|        |                       | service_state_table start at 0x0000 |
| 0x0000 | running               | unit is ~                           |
| 0x0001 | dead                  | unit is ~                           |
| 0x0002 | start-pre             | unit is ~                           |
| 0x0003 | start                 | unit is ~                           |
| 0x0004 | exited                | unit is ~                           |
| 0x0005 | reload                | unit is ~                           |
| 0x0006 | stop                  | unit is ~                           |
| 0x0007 | stop-watchdog         | unit is ~                           |
| 0x0008 | stop-sigterm          | unit is ~                           |
| 0x0009 | stop-sigkill          | unit is ~                           |
| 0x000a | stop-post             | unit is ~                           |
| 0x000b | final-sigterm         | unit is ~                           |
| 0x000c | failed                | unit is ~                           |
| 0x000d | auto-restart          | unit is ~                           |
|        |                       | service_state_table start at 0x0010 |
| 0x0010 | waiting               | unit is ~                           |
|        |                       | service_state_table start at 0x0020 |
| 0x0020 | tentative             | unit is ~                           |
| 0x0021 | plugged               | unit is ~                           |
|        |                       | service_state_table start at 0x0030 |
| 0x0030 | mounting              | unit is ~                           |
| 0x0031 | mounting-done         | unit is ~                           |
| 0x0032 | mounted               | unit is ~                           |
| 0x0033 | remounting            | unit is ~                           |
| 0x0034 | unmounting            | unit is ~                           |
| 0x0035 | remounting-sigterm    | unit is ~                           |
| 0x0036 | remounting-sigkill    | unit is ~                           |
| 0x0037 | unmounting-sigterm    | unit is ~                           |
| 0x0038 | unmounting-sigkill    | unit is ~                           |
|        |                       | service_state_table start at 0x0040 |
|        |                       | service_state_table start at 0x0050 |
| 0x0050 | abandoned             | unit is ~                           |
|        |                       | service_state_table start at 0x0060 |
| 0x0060 | active                | unit is ~                           |
|        |                       | service_state_table start at 0x0070 |
| 0x0070 | start-chown           | unit is ~                           |
| 0x0071 | start-post            | unit is ~                           |
| 0x0072 | listening             | unit is ~                           |
| 0x0073 | stop-pre              | unit is ~                           |
| 0x0074 | stop-pre-sigterm      | unit is ~                           |
| 0x0075 | stop-pre-sigkill      | unit is ~                           |
| 0x0076 | final-sigkill         | unit is ~                           |
|        |                       | service_state_table start at 0x0080 |
| 0x0080 | activating            | unit is ~                           |
| 0x0081 | activating-done       | unit is ~                           |
| 0x0082 | deactivating          | unit is ~                           |
| 0x0083 | deactivating-sigterm  | unit is ~                           |
| 0x0084 | deactivating-sigkill  | unit is ~                           |
|        |                       | service_state_table start at 0x0090 |
|        |                       | service_state_table start at 0x00a0 |
| 0x00a0 | elapsed               | unit is ~                           |
|        |                       |                                     |
| -1     | err                   | field not found in lookup table     |

### Example Output

Linux Systemd Units:
```
$ telegraf --test --config /tmp/telegraf.conf
> systemd_units,host=host1.example.com,name=dbus.service load=0i,active=0i,sub=0i 1533730725000000000
> systemd_units,host=host1.example.com,name=networking.service load=0i,active=3i,sub=12i 1533730725000000000
> systemd_units,host=host1.example.com,name=ssh.service load=0i,active=0i,sub=0i 1533730725000000000
...
```

### Possible Improvements
- add blacklist to filter names
