# mdstat Input Plugin

The mdstat plugin gathers statistics about any Linux MD RAID arrays configured on the host
by reading /proc/mdstat. For a full list of available fields see the
/proc/mdstat section of the [proc man page](http://man7.org/linux/man-pages/man5/proc.5.html).
For a better idea of what each field represents, see the
[mdstat man page](https://raid.wiki.kernel.org/index.php/Mdstat).

Stat collection based on Prometheus' mdstat collection library at <https://github.com/prometheus/procfs/blob/master/mdstat.go>

## Configuration

```toml
# Get kernel statistics from /proc/mdstat
[[inputs.mdstat]]
  ## Sets file path
  ## If not specified, then default is /proc/mdstat
  # file_name = "/proc/mdstat"
```

## Measurements & Fields

- mdstat
  - BlocksSynced (if the array is rebuilding/checking, this is the count of blocks that have been scanned)
  - BlocksSyncedFinishTime (the expected finish time of the rebuild scan, listed in minutes remaining)
  - BlocksSyncedPct (the percentage of the rebuild scan left)
  - BlocksSyncedSpeed (the current speed the rebuild is running at, listed in K/sec)
  - BlocksTotal (the total count of blocks in the array)
  - DisksActive (the number of disks that are currently considered healthy in the array)
  - DisksFailed (the current count of failed disks in the array)
  - DisksSpare (the current count of "spare" disks in the array)
  - DisksTotal (total count of disks in the array)

## Tags

- mdstat
  - ActivityState (`active` or `inactive`)
  - Devices (comma separated list of devices that make up the array)
  - Name (name of the array)

## Example Output

```shell
$ telegraf --config ~/ws/telegraf.conf --input-filter mdstat --test
* Plugin: mdstat, Collection 1
> mdstat,ActivityState=active,Devices=sdm1\,sdn1,Name=md1 BlocksSynced=231299072i,BlocksSyncedFinishTime=0,BlocksSyncedPct=0,BlocksSyncedSpeed=0,BlocksTotal=231299072i,DisksActive=2i,DisksFailed=0i,DisksSpare=0i,DisksTotal=2i,DisksDown=0i 1617814276000000000
> mdstat,ActivityState=active,Devices=sdm5\,sdn5,Name=md2 BlocksSynced=2996224i,BlocksSyncedFinishTime=0,BlocksSyncedPct=0,BlocksSyncedSpeed=0,BlocksTotal=2996224i,DisksActive=2i,DisksFailed=0i,DisksSpare=0i,DisksTotal=2i,DisksDown=0i 1617814276000000000
```
