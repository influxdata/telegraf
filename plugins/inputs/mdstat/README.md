# MDSTAT Input Plugin

The `mdstat` plugin gathers metrics on raid arrays on the system that are managed by mdadm

#### Configuration
```toml
[[inputs.mdstat]]
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  # host_proc = "/proc"
```

Additionally, if the `host_proc` config item is not set, the plugin will try to get the
location of proc directory from the `HOST_PROC` environment variable. Failing this,
it will default to `/proc`.

### Metrics

- mdstat_device
  - tags
    - device
  - fields
    - status(string)
    - raidType (string)
    - minDisks (raid)
    - currDisks (int)
    - missingDisks (int)
    - failedDisks (int)
    - inRecovery (bool)
    - recoveryPercent (float)
- mdstat_disk
  - tags
    - device
    - disk
  - fields
    - role (int)
    - failed (bool)

### Example Output
```
mdstat_device,device=md0,host=grommit currDisks=4i,failedDisks=0i,inRecovery=false,minDisks=4i,missingDisks=0i,raidType="raid5",recoveryPercent=0,status="active" 1616607748000000000
mdstat_disk,device=md0,disk=sdb,host=grommit failed=false,role=1i 1616607748000000000
mdstat_disk,device=md0,disk=sda,host=grommit failed=false,role=0i 1616607748000000000
mdstat_disk,device=md0,disk=sdc,host=grommit failed=false,role=2i 1616607748000000000
mdstat_disk,device=md0,disk=sdd,host=grommit failed=false,role=4i 1616607748000000000
```
