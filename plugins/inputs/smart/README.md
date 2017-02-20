# Telegraf S.M.A.R.T. plugin

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

## Measurements

- smart_device:

    * Tags:
      - `capacity`
      - `device`
      - `device_model`
      - `enabled`
      - `health`
      - `serial_no`
    * Fields:
      - `exit_status`

- smart_attribute:

    * Tags:
      - `device`
      - `fail`
      - `flags`
      - `id`
      - `name`
    * Fields:
      - `exit_status`
      - `raw_value`
      - `threshold`
      - `value`
      - `worst`

### Flags

The interpretation of the tag `flags` is:
 - *K* auto-keep
 - *C* event count
 - *R* error rate
 - *S* speed/performance
 - *O* updated online
 - *P* prefailure warning

## Configuration

```toml
# Read metrics from storage devices supporting S.M.A.R.T.
[[inputs.smart]]
  ## Optionally specify the path to the smartctl executable
  # path = "/usr/bin/smartctl"
  #
  ## Skip checking disks in this power mode. Defaults to
  ## "standby" to not wake up disks that have stoped rotating.
  ## See --nockeck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of
  ## power mode and might require changing this value to
  ## "never" depending on your storage device.
  # nocheck = "standby"
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

To run `smartctl` with `sudo` create a wrapper script and use `path` in
the configuration to execute that.

## Output

Example output from an _Apple SSD_:
```
> smart_attribute,device=/dev/rdisk0,id=194,name=Temperature_Celsius,flags=-O---K,fail=-,host=STIZ0039.lan exit_status=0i,value=64i,worst=21i,threshold=0i,raw_value=36i 1487632495000000000
> smart_attribute,device=/dev/rdisk0,id=197,name=Current_Pending_Sector,flags=-O---K,fail=-,host=STIZ0039.lan exit_status=0i,value=100i,worst=100i,threshold=0i,raw_value=0i 1487632495000000000
> smart_attribute,device=/dev/rdisk0,id=199,name=UDMA_CRC_Error_Count,flags=-O-RC-,fail=-,host=STIZ0039.lan exit_status=0i,value=200i,worst=200i,threshold=0i,raw_value=0i 1487632495000000000
> smart_device,device_model=APPLE\ SSD\ SM256E,serial_no=S0X5NZBC422720,capacity=251000193024,enabled=Enabled,health=PASSED,host=STIZ0039.lan,device=/dev/rdisk0 exit_status=0i 1487632495000000000
```
