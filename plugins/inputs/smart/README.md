# S.M.A.R.T. Input Plugin

Get metrics using the command line utility `smartctl` for S.M.A.R.T. (Self-Monitoring, Analysis and Reporting Technology) storage devices. SMART is a monitoring system included in computer hard disk drives (HDDs) and solid-state drives (SSDs)[1] that detects and reports on various indicators of drive reliability, with the intent of enabling the anticipation of hardware failures.
See smartmontools (https://www.smartmontools.org/).

If no devices are specified, the plugin will scan for SMART devices via the following command:

```
smartctl --scan
```

Metrics will be reported from the following `smartctl` command:

```
smartctl --info --attributes --health -n <nocheck> --format=brief <device>
```

This plugin supports _smartmontools_ version 5.41 and above, but v. 5.41 and v. 5.42
might require setting `nocheck`, see the comment in the sample configuration.

To enable SMART on a storage device run:

```
smartctl -s on <device>
```

### Configuration:

```toml
# Read metrics from storage devices supporting S.M.A.R.T.
[[inputs.smart]]
  ## Optionally specify the path to the smartctl executable
  # path = "/usr/bin/smartctl"
  #
  ## On most platforms smartctl requires root access.
  ## Setting 'use_sudo' to true will make use of sudo to run smartctl.
  ## Sudo must be configured to to allow the telegraf user to run smartctl
  ## with out password.
  # use_sudo = false
  #
  ## Skip checking disks in this power mode. Defaults to
  ## "standby" to not wake up disks that have stoped rotating.
  ## See --nockeck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of
  ## power mode and might require changing this value to
  ## "never" depending on your storage device.
  # nocheck = "standby"
  #
  ## Gather detailed metrics for each SMART Attribute.
  ## Defaults to "false"
  ##
  # attributes = false
  #
  ## Optionally specify devices to exclude from reporting.
  # excludes = [ "/dev/pass6" ]
  #
  ## Optionally specify devices and device type, if unset
  ## a scan (smartctl --scan) for S.M.A.R.T. devices will
  ## done and all found will be included except for the
  ## excluded in excludes.
  # devices = [ "/dev/ada0 -d atacam" ]
```

### Metrics:

- smart_device:
  - tags:
    - capacity
    - device
    - device_model
    - enabled
    - health
    - serial_no
    - wwn
  - fields:
    - exit_status
    - health_ok
    - read_error_rate
    - seek_error
    - temp_c
    - udma_crc_errors

- smart_attribute:
  - tags:
    - device
    - fail
    - flags
    - id
    - name
    - serial_no
    - wwn
  - fields:
    - exit_status
    - raw_value
    - threshold
    - value
    - worst

#### Flags

The interpretation of the tag `flags` is:
 - `K` auto-keep
 - `C` event count
 - `R` error rate
 - `S` speed/performance
 - `O` updated online
 - `P` prefailure warning

#### Exit Status

The `exit_status` field captures the exit status of the smartctl command which
is defined by a bitmask. For the interpretation of the bitmask see the man page for
smartctl.

#### Device Names

Device names, e.g., `/dev/sda`, are *not persistent*, and may be
subject to change across reboots or system changes. Instead, you can the
*World Wide Name* (WWN) or serial number to identify devices. On Linux block
devices can be referenced by the WWN in the following location:
`/dev/disk/by-id/`.

To run `smartctl` with `sudo` create a wrapper script and use `path` in
the configuration to execute that.

### Output

```
smart_device,enabled=Enabled,host=mbpro.local,device=rdisk0,model=APPLE\ SSD\ SM0512F,serial_no=S1K5NYCD964433,wwn=5002538655584d30,capacity=500277790720 udma_crc_errors=0i,exit_status=0i,health_ok=true,read_error_rate=0i,temp_c=40i 1502536854000000000
smart_attribute,serial_no=S1K5NYCD964433,wwn=5002538655584d30,id=199,name=UDMA_CRC_Error_Count,flags=-O-RC-,fail=-,host=mbpro.local,device=rdisk0 threshold=0i,raw_value=0i,exit_status=0i,value=200i,worst=200i 1502536854000000000
smart_attribute,device=rdisk0,serial_no=S1K5NYCD964433,wwn=5002538655584d30,id=240,name=Unknown_SSD_Attribute,flags=-O---K,fail=-,host=mbpro.local exit_status=0i,value=100i,worst=100i,threshold=0i,raw_value=0i 1502536854000000000
```
