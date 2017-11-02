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

`docker run -v /:/hostfs:ro -e HOST_MOUNT_PREFIX=/hostfs -e HOST_ETC=/hostfs/etc telegraf`

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
    - mode (whether the mount is rw or ro)

### Example Output:

```
% ./telegraf --config ~/ws/telegraf.conf --input-filter disk --test
* Plugin: disk, Collection 1
> disk,fstype=hfs,mode=ro,path=/ free=398407520256i,inodes_free=97267461i,inodes_total=121847806i,inodes_used=24580345i,total=499088621568i,used=100418957312i,used_percent=20.131039916242397 1453832006274071563
> disk,fstype=devfs,mode=rw,path=/dev free=0i,inodes_free=0i,inodes_total=628i,inodes_used=628i,total=185856i,used=185856i,used_percent=100 1453832006274137913
> disk,fstype=autofs,mode=rw,path=/net free=0i,inodes_free=0i,inodes_total=0i,inodes_used=0i,total=0i,used=0i,used_percent=0 1453832006274157077
> disk,fstype=autofs,mode=rw,path=/home free=0i,inodes_free=0i,inodes_total=0i,inodes_used=0i,total=0i,used=0i,used_percent=0 1453832006274169688
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
    - read_bytes (integer, counter, bytes)
    - write_bytes (integer, counter, bytes)
    - read_time (integer, counter, milliseconds)
    - write_time (integer, counter, milliseconds)
    - io_time (integer, counter, milliseconds)
    - weighted_io_time (integer, counter, milliseconds)
    - iops_in_progress (integer, gauge)

On linux these values correspond to the values in [`/proc/diskstats`](https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats) and [`/sys/block/<dev>/stat`](https://www.kernel.org/doc/Documentation/block/stat.txt).

#### `reads` & `writes`:

These values increment when an I/O request completes.

#### `read_bytes` & `write_bytes`:

These values count the number of bytes read from or written to this
block device.

#### `read_time` & `write_time`:

These values count the number of milliseconds that I/O requests have
waited on this block device.  If there are multiple I/O requests waiting,
these values will increase at a rate greater than 1000/second; for
example, if 60 read requests wait for an average of 30 ms, the read_time
field will increase by 60*30 = 1800.

#### `io_time`:

This value counts the number of milliseconds during which the device has
had I/O requests queued.

#### `weighted_io_time`:

This value counts the number of milliseconds that I/O requests have waited
on this block device.  If there are multiple I/O requests waiting, this
value will increase as the product of the number of milliseconds times the
number of requests waiting (see `read_time` above for an example).

#### `iops_in_progress`:

This value counts the number of I/O requests that have been issued to
the device driver but have not yet completed.  It does not include I/O
requests that are in the queue but not yet issued to the device driver.

### Tags:

- All measurements have the following tags:
    - name (device name)
- If configured to use serial numbers (default: disabled):
    - serial (device serial number)

### Sample Queries:

#### Calculate percent IO utilization per disk and host:
```
SELECT derivative(last("io_time"),1ms) FROM "diskio" WHERE time > now() - 30m GROUP BY "host","name",time(60s)
```

#### Calculate average queue depth:
`iops_in_progress` will give you an instantaneous value. This will give you the average between polling intervals.
```
SELECT derivative(last("weighted_io_time",1ms)) from "diskio" WHERE time > now() - 30m GROUP BY "host","name",time(60s)
```

### Example Output:

```
% telegraf -config ~/.telegraf/telegraf.conf -input-filter diskio -test
* Plugin: inputs.diskio, Collection 1
> diskio,name=sda weighted_io_time=8411917i,read_time=7446444i,write_time=971489i,io_time=866197i,write_bytes=5397686272i,iops_in_progress=0i,reads=2970519i,writes=361139i,read_bytes=119528903168i 1502467254359000000
> diskio,name=sda1 reads=2149i,read_bytes=10753536i,write_bytes=20697088i,write_time=346i,weighted_io_time=505i,writes=2110i,read_time=161i,io_time=208i,iops_in_progress=0i 1502467254359000000
> diskio,name=sda2 reads=2968279i,writes=359029i,write_bytes=5376989184i,iops_in_progress=0i,weighted_io_time=8411250i,read_bytes=119517334528i,read_time=7446249i,write_time=971143i,io_time=866010i 1502467254359000000
> diskio,name=sdb writes=99391856i,write_time=466700894i,io_time=630259874i,weighted_io_time=4245949844i,reads=2750773828i,read_bytes=80667939499008i,write_bytes=6329347096576i,read_time=3783042534i,iops_in_progress=2i 1502467254359000000
> diskio,name=centos/root read_time=7472461i,write_time=950014i,iops_in_progress=0i,weighted_io_time=8424447i,writes=298543i,read_bytes=119510105088i,io_time=837421i,reads=2971769i,write_bytes=5192795648i 1502467254359000000
> diskio,name=centos/var_log reads=1065i,writes=69711i,read_time=1083i,write_time=35376i,read_bytes=6828032i,write_bytes=184193536i,io_time=29699i,iops_in_progress=0i,weighted_io_time=36460i 1502467254359000000
> diskio,name=postgresql/pgsql write_time=478267417i,io_time=631098730i,iops_in_progress=2i,weighted_io_time=4263637564i,reads=2750777151i,writes=110044361i,read_bytes=80667939288064i,write_bytes=6329347096576i,read_time=3784499336i 1502467254359000000
```
