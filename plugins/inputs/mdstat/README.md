# mdstat Input Plugin

The mdstat plugin gathers statistics about any Linux MD RAID arrays configured
on the host by reading /proc/mdstat. For a full list of available fields see
the /proc/mdstat section of the [proc man page][man-proc].  For a better idea
of what each field represents, see the [mdstat man page][man-mdstat].

Stat collection based on Prometheus' [mdstat collection library][prom-lib].

[man-proc]: http://man7.org/linux/man-pages/man5/proc.5.html

[man-mdstat]: https://raid.wiki.kernel.org/index.php/Mdstat

[prom-lib]: https://github.com/prometheus/procfs/blob/master/mdstat.go

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Get kernel statistics from /proc/mdstat
# This plugin ONLY supports Linux
[[inputs.mdstat]]
  ## Sets file path
  ## If not specified, then default is /proc/mdstat
  # file_name = "/proc/mdstat"
```

## Metrics

- mdstat
  - BlocksSynced (if the array is rebuilding/checking, this is the count of
    blocks that have been scanned)
  - BlocksSyncedFinishTime (the expected finish time of the rebuild scan,
    listed in minutes remaining)
  - BlocksSyncedPct (the percentage of the rebuild scan left)
  - BlocksSyncedSpeed (the current speed the rebuild is running at, listed
    in K/sec)
  - BlocksTotal (the total count of blocks in the array)
  - DisksActive (the number of disks that are currently considered healthy
    in the array)
  - DisksFailed (the current count of failed disks in the array)
  - DisksSpare (the current count of "spare" disks in the array)
  - DisksTotal (total count of disks in the array)

## Tags

- mdstat
  - ActivityState (`active` or `inactive`)
  - Devices (comma separated list of devices that make up the array)
  - Name (name of the array)

## Example Output

```text
mdstat,ActivityState=active,Devices=sdm1\,sdn1,Name=md1 BlocksSynced=231299072i,BlocksSyncedFinishTime=0,BlocksSyncedPct=0,BlocksSyncedSpeed=0,BlocksTotal=231299072i,DisksActive=2i,DisksFailed=0i,DisksSpare=0i,DisksTotal=2i,DisksDown=0i 1617814276000000000
mdstat,ActivityState=active,Devices=sdm5\,sdn5,Name=md2 BlocksSynced=2996224i,BlocksSyncedFinishTime=0,BlocksSyncedPct=0,BlocksSyncedSpeed=0,BlocksTotal=2996224i,DisksActive=2i,DisksFailed=0i,DisksSpare=0i,DisksTotal=2i,DisksDown=0i 1617814276000000000
```
