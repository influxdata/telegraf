# DiskIO Input Plugin

The diskio input plugin gathers metrics about disk traffic and timing.

### Configuration:

```toml
# Read metrics about disk IO by device
[[inputs.diskio]]
  ## By default, telegraf will gather stats for all devices including
  ## disk partitions.
  ## Setting devices will restrict the stats to the specified devices.
  # devices = ["sda", "sdb"]
  ## Uncomment the following line if you need disk serial numbers.
  # skip_serial_number = false
  #
  ## On systems which support it, device metadata can be added in the form of
  ## tags.
  ## Currently only Linux is supported via udev properties. You can view
  ## available properties for a device by running:
  ## 'udevadm info -q property -n /dev/sda'
  ## Note: Most, but not all, udev properties can be accessed this way. Properties
  ## that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
  # device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]
  #
  ## Using the same metadata source as device_tags, you can also customize the
  ## name of the device via templates.
  ## The 'name_templates' parameter is a list of templates to try and apply to
  ## the device. The template may contain variables in the form of '$PROPERTY' or
  ## '${PROPERTY}'. The first template which does not contain any variables not
  ## present for the device is used as the device name tag.
  ## The typical use case is for LVM volumes, to get the VG/LV name instead of
  ## the near-meaningless DM-0 name.
  # name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME"]
```

#### Docker container

To monitor the Docker engine host from within a container you will need to
mount the host's filesystem into the container and set the `HOST_PROC`
environment variable to the location of the `/proc` filesystem.  Additionally,
it is required to use privileged mode to provide access to `/dev`.

If you are using the `device_tags` or `name_templates` options, you will need
to bind mount `/run/udev` into the container.

```
docker run --privileged -v /:/hostfs:ro -v /run/udev:/run/udev:ro -e HOST_PROC=/hostfs/proc telegraf
```

### Metrics:

- diskio
  - tags:
    - name (device name)
    - serial (device serial number)
  - fields:
    - reads (integer, counter)
    - writes (integer, counter)
    - read_bytes (integer, counter, bytes)
    - write_bytes (integer, counter, bytes)
    - read_time (integer, counter, milliseconds)
    - write_time (integer, counter, milliseconds)
    - io_time (integer, counter, milliseconds)
    - weighted_io_time (integer, counter, milliseconds)
    - iops_in_progress (integer, gauge)

On linux these values correspond to the values in
[`/proc/diskstats`](https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats)
and
[`/sys/block/<dev>/stat`](https://www.kernel.org/doc/Documentation/block/stat.txt).

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

### Sample Queries:

#### Calculate percent IO utilization per disk and host:
```
SELECT non_negative_derivative(last("io_time"),1ms) FROM "diskio" WHERE time > now() - 30m GROUP BY "host","name",time(60s)
```

#### Calculate average queue depth:
`iops_in_progress` will give you an instantaneous value. This will give you the average between polling intervals.
```
SELECT non_negative_derivative(last("weighted_io_time",1ms)) from "diskio" WHERE time > now() - 30m GROUP BY "host","name",time(60s)
```

### Example Output:

```
diskio,name=sda weighted_io_time=8411917i,read_time=7446444i,write_time=971489i,io_time=866197i,write_bytes=5397686272i,iops_in_progress=0i,reads=2970519i,writes=361139i,read_bytes=119528903168i 1502467254359000000
diskio,name=sda1 reads=2149i,read_bytes=10753536i,write_bytes=20697088i,write_time=346i,weighted_io_time=505i,writes=2110i,read_time=161i,io_time=208i,iops_in_progress=0i 1502467254359000000
diskio,name=sda2 reads=2968279i,writes=359029i,write_bytes=5376989184i,iops_in_progress=0i,weighted_io_time=8411250i,read_bytes=119517334528i,read_time=7446249i,write_time=971143i,io_time=866010i 1502467254359000000
diskio,name=sdb writes=99391856i,write_time=466700894i,io_time=630259874i,weighted_io_time=4245949844i,reads=2750773828i,read_bytes=80667939499008i,write_bytes=6329347096576i,read_time=3783042534i,iops_in_progress=2i 1502467254359000000
diskio,name=centos/root read_time=7472461i,write_time=950014i,iops_in_progress=0i,weighted_io_time=8424447i,writes=298543i,read_bytes=119510105088i,io_time=837421i,reads=2971769i,write_bytes=5192795648i 1502467254359000000
diskio,name=centos/var_log reads=1065i,writes=69711i,read_time=1083i,write_time=35376i,read_bytes=6828032i,write_bytes=184193536i,io_time=29699i,iops_in_progress=0i,weighted_io_time=36460i 1502467254359000000
diskio,name=postgresql/pgsql write_time=478267417i,io_time=631098730i,iops_in_progress=2i,weighted_io_time=4263637564i,reads=2750777151i,writes=110044361i,read_bytes=80667939288064i,write_bytes=6329347096576i,read_time=3784499336i 1502467254359000000
```
