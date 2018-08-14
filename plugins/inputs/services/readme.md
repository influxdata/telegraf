# Services Plugin

The services plugin gathers services status for Linux systemd units and
Windows services. On Linux, it relies on ```systemctl list-units --type=service```
where on Windows it uses [WMI Scripting](https://docs.microsoft.com/en-us/windows/desktop/wmisdk/scripting-api-objects)
to collect data on autostart services.

The results are tagged with the service name and provide a status field
indicating when an autostart service failed.

### Configuration
```
[[inputs.services]]
  ## The default timeout of 1s for systemctl execution can be overridden here:
  # timeout = "1s"
```

### Metrics
- services
  - tags:
    - name (string, service name)
  - fields:
    - state (string, systemd active/sub fields or win32_services.state field)
    - status (int, nagios-style simple status, see below)

#### Statuses

| Value | Meaning  | Description                                       |
| ----- | -------  | -----------                                       |
| 0     | Ok       | Service state is without failure                  |
| 1     | Warning  | Service state indicates pre-fault condition       |
| 2     | Critical | Service state indicates failure                   |
| 3     | Unknown  | Service state did not match expecations           |

Note: values are identical in their meaning to other monitoring solutions like Nagios.

### Example Output

Linux Systemd Units:
```
$ telegraf --test --config /tmp/telegraf.conf
> services,host=host1.example.com,name=dbus.service state="active/running",status=0i 1533730725000000000
> services,host=host1.example.com,name=networking.service state="failed/failed",status=2i 1533730725000000000
> services,host=host1.example.com,name=ssh.service state="active/running",status=0i 1533730725000000000
...
```

Windows Services:
```
$ telegraf --test --config c:\temp\telegraf.conf
> services,host=host2.example.com,name=cplspcon state="Running",status=0i 1533731001000000000
> services,host=host2.example.com,name=CryptSvc state="Running",status=0i 1533731001000000000
> services,host=host2.example.com,name=dbupdate state="Stopped",status=2i 1533731001000000000
...
```

### Possible Improvements
- add timeout for wmi calls
- add blacklist to filter names
