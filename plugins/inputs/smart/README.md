# Telegraf S.M.A.R.T. plugin

Get metrics using the command line utility `smartctl` for S.M.A.R.T. (Self-Monitoring, Analysis and Reporting Technology) storage devices. SMART is a monitoring system included in computer hard disk drives (HDDs) and solid-state drives (SSDs)[1] that detects and reports on various indicators of drive reliability, with the intent of enabling the anticipation of hardware failures.
See smartmontools(https://www.smartmontools.org/).

If no devices are specified, the plugin will scan for SMART devices via the following command:

```
smartctl --scan
```

On some platforms (e.g. Darwin/macOS) this doesn't return a useful list of devices and you must instead specify which devices to collect metrics from in the configuration file.

Metrics will be reported from the following `smartctl` command:

```
smartctl --info --attributes --nocheck=standby --format=brief <device>
```

## Measurements

- smart:

    * Tags:
      - `device`
      - `device_model`
      - `serial_no`
      - `capacity`
      - `enabled`
      - `id`
      - `name`
      - `flags`
      - `fail`
    * Fields:
      - `value`
      - `worst`
      - `threshold`
      - `raw_value`

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
  ## optionally specify the path to the smartctl executable
  # path = "/usr/bin/smartctl"
  #
  ## optionally specify devices to exclude from reporting.
  # exclude = [ "/dev/pass6" ]
  #
  ## optionally specify devices, if unset all S.M.A.R.T. devices
  ## will be included
  # devices = [ "/dev/ada0" ]
```

## Output

When retrieving stats from the local machine (no server specified):
```
> smart,serial_no=WD-WMC4N0900000,id=1,name=Raw_Read_Error_Rate,flags=POSR-K,fail=-,host=example,device=/dev/ada0,device_model=WDC\ WD30EFRX-68EUZN0,capacity=3000592982016,enabled=Enabled value=200i,worst=200i,threshold=51i,raw_value=0i 1486892929000000000
> smart,serial_no=WD-WMC4N0900000,device=/dev/ada0,device_model=WDC\ WD30EFRX-68EUZN0,capacity=3000592982016,enabled=Enabled,id=3,name=Spin_Up_Time,flags=POS--K,fail=-,host=example value=181i,worst=180i,threshold=21i,raw_value=5916i 1486892929000000000
> smart,device_model=WDC\ WD30EFRX-68EUZN0,capacity=3000592982016,enabled=Enabled,name=Start_Stop_Count,flags=-O--CK,fail=-,device=/dev/ada0,serial_no=WD-WMC4N0900000,id=4,host=example value=100i,worst=100i,threshold=0i,raw_value=18i 1486892929000000000
> smart,enabled=Enabled,device_model=WDC\ WD30EFRX-68EUZN0,id=5,name=Reallocated_Sector_Ct,capacity=3000592982016,device=/dev/ada0,serial_no=WD-WMC4N0900000,flags=PO--CK,fail=-,host=example value=200i,worst=200i,threshold=140i,raw_value=0i 1486892929000000000
> smart,serial_no=WD-WMC4N0900000,capacity=3000592982016,enabled=Enabled,name=Seek_Error_Rate,host=example,device=/dev/ada0,id=7,flags=-OSR-K,fail=-,device_model=WDC\ WD30EFRX-68EUZN0 value=200i,worst=200i,threshold=0i,raw_value=0i 1486892929000000000
> smart,flags=-O--CK,device_model=WDC\ WD30EFRX-68EUZN0,capacity=3000592982016,enabled=Enabled,id=9,name=Power_On_Hours,fail=-,host=example,device=/dev/ada0,serial_no=WD-WMC4N0900000 value=65i,worst=65i,threshold=0i,raw_value=25998i 1486892929000000000
```
