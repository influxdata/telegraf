# Disk Input Plugin

The disk input plugin gathers metrics about disk usage.

Note that `used_percent` is calculated by doing `used / (used + free)`, _not_
`used / total`, which is how the unix `df` command does it. See
https://en.wikipedia.org/wiki/Df_(Unix) for more details.

### Configuration:

```
# Read metrics about disk usage by mount point
[[inputs.disk]]
  # By default, telegraf gather stats for all mountpoints.
  # Setting mountpoints will restrict the stats to the specified mountpoints.
  # mount_points = ["/"]
```

Additionally, the behavior of resolving the `mount_points` can be configured by using the `HOST_MOUNT_PREFIX` environment variable.
When present, this variable is prepended to the mountpoints discovered by the plugin before retrieving stats.
The prefix is stripped from the reported `path` in the measurement.
This settings is useful when running `telegraf` inside a docker container to report host machine metrics.
In this case, the host's root volume should be mounted into the container and the `HOST_MOUNT_PREFIX` and `HOST_ETC` environment variables set.

`docker run -v /:/hostfs:ro -e HOST_MOUNT_PREFIX=/hostfs -e HOST_ETC=/hostfs/etc telegraf-docker`

### Measurements & Fields:

- disk
    - free (integer, bytes)
    - total (integer, bytes)
    - used (integer, bytes)
    - used_percent (float, percent)
    - inodes_free (integer, files)
    - inodes_total (integer, files)
    - inodes_used (integer, files)

### Tags:

- All measurements have the following tags:
    - fstype (filesystem type)
    - path (mount point path)

### Example Output:

```
% ./telegraf -config ~/ws/telegraf.conf -input-filter disk -test
* Plugin: disk, Collection 1
> disk,fstype=hfs,path=/ free=398407520256i,inodes_free=97267461i,inodes_total=121847806i,inodes_used=24580345i,total=499088621568i,used=100418957312i,used_percent=20.131039916242397 1453832006274071563
> disk,fstype=devfs,path=/dev free=0i,inodes_free=0i,inodes_total=628i,inodes_used=628i,total=185856i,used=185856i,used_percent=100 1453832006274137913
> disk,fstype=autofs,path=/net free=0i,inodes_free=0i,inodes_total=0i,inodes_used=0i,total=0i,used=0i,used_percent=0 1453832006274157077
> disk,fstype=autofs,path=/home free=0i,inodes_free=0i,inodes_total=0i,inodes_used=0i,total=0i,used=0i,used_percent=0 1453832006274169688
```


# DiskIO Input Plugin

The diskio input plugin gathers metrics about disk traffic and timing.

### Configuration:

```
# Read metrics about disk IO by device
[[inputs.diskio]]
  ## By default, telegraf will gather stats for all devices including
  ## disk partitions.
  ## Setting devices will restrict the stats to the specified devices.
  # devices = ["sda", "sdb"]
  ## Uncomment the following line if you need disk serial numbers.
  # skip_serial_number = false
```

Data collection is based on github.com/shirou/gopsutil. This package handles platform dependencies and converts all timing information to milliseconds.


### Measurements & Fields:

- diskio
    - reads (integer, counter)
    - writes (integer, counter)
    - read_bytes (integer, bytes)
    - write_bytes (integer, bytes)
    - read_time (integer, milliseconds)
    - write_time (integer, milliseconds)
    - io_time (integer, milliseconds)
    - iops_in_progress (integer, counter) (since #2037, not yet in STABLE)

### Tags:

- All measurements have the following tags:
    - name (device name)
- If configured to use serial numbers (default: disabled):
    - serial (device serial number)

### Example Output:

```
% telegraf -config ~/.telegraf/telegraf.conf -input-filter diskio -test
* Plugin: inputs.diskio, Collection 1
> diskio,name=mmcblk1p2 io_time=244i,read_bytes=966656i,read_time=276i,reads=128i,write_bytes=0i,write_time=0i,writes=0i 1484916036000000000
> diskio,name=mmcblk1boot1 io_time=264i,read_bytes=90112i,read_time=264i,reads=22i,write_bytes=0i,write_time=0i,writes=0i 1484916036000000000
> diskio,name=mmcblk1boot0 io_time=212i,read_bytes=90112i,read_time=212i,reads=22i,write_bytes=0i,write_time=0i,writes=0i 1484916036000000000
> diskio,name=mmcblk0 io_time=1855380i,read_bytes=135861248i,read_time=58484i,reads=4081i,write_bytes=364068864i,write_time=7128792i,writes=18019i 1484916036000000000
> diskio,name=mmcblk0p1 io_time=1855256i,read_bytes=134915072i,read_time=58256i,reads=3958i,write_bytes=364068864i,write_time=7128792i,writes=18019i 1484916036000000000
> diskio,name=mmcblk1 io_time=384i,read_bytes=2633728i,read_time=728i,reads=323i,write_bytes=0i,write_time=0i,writes=0i 1484916036000000000
> diskio,name=mmcblk1p1 io_time=216i,read_bytes=860160i,read_time=288i,reads=106i,write_bytes=0i,write_time=0i,writes=0i 1484916036000000000
```
