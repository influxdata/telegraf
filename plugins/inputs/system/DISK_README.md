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
