# Disk Input Plugin

The disk input plugin gathers metrics about disk usage.

Note that `used_percent` is calculated by doing `used / (used + free)`, _not_
`used / total`, which is how the unix `df` command does it. See
[wikipedia - df](https://en.wikipedia.org/wiki/Df_(Unix)) for more details.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about disk usage by mount point
[[inputs.disk]]
  ## By default stats will be gathered for all mount points.
  ## Set mount_points will restrict the stats to only the specified mount points.
  # mount_points = ["/"]

  ## Ignore mount points by filesystem type.
  ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

  ## Ignore mount points by mount options.
  ## The 'mount' command reports options of all mounts in parathesis.
  ## Bind mounts can be ignored with the special 'bind' option.
  # ignore_mount_opts = []
```

### Docker container

To monitor the Docker engine host from within a container you will need to mount
the host's filesystem into the container and set the `HOST_PROC` environment
variable to the location of the `/proc` filesystem.  If desired, you can also
set the `HOST_MOUNT_PREFIX` environment variable to the prefix containing the
`/proc` directory, when present this variable is stripped from the reported
`path` tag.

```shell
docker run -v /:/hostfs:ro -e HOST_MOUNT_PREFIX=/hostfs -e HOST_PROC=/hostfs/proc telegraf
```

## Metrics

- disk
  - tags:
    - fstype (filesystem type)
    - device (device file)
    - path (mount point path)
    - mode (whether the mount is rw or ro)
    - label (devicemapper labels, only if present)
  - fields:
    - free (integer, bytes)
    - total (integer, bytes)
    - used (integer, bytes)
    - used_percent (float, percent)
    - inodes_free (integer, files)
    - inodes_total (integer, files)
    - inodes_used (integer, files)

## Troubleshooting

On Linux, the list of disks is taken from the `/proc/self/mounts` file and a
[statfs] call is made on the second column.  If any expected filesystems are
missing ensure that the `telegraf` user can read these files:

```shell
$ sudo -u telegraf cat /proc/self/mounts | grep sda2
/dev/sda2 /home ext4 rw,relatime,data=ordered 0 0
$ sudo -u telegraf stat /home
```

It may be desired to use POSIX ACLs to provide additional access:

```shell
sudo setfacl -R -m u:telegraf:X /var/lib/docker/volumes/
```

## Example Output

```text
disk,fstype=hfs,mode=ro,path=/ free=398407520256i,inodes_free=97267461i,inodes_total=121847806i,inodes_used=24580345i,total=499088621568i,used=100418957312i,used_percent=20.131039916242397 1453832006274071563
disk,fstype=devfs,mode=rw,path=/dev free=0i,inodes_free=0i,inodes_total=628i,inodes_used=628i,total=185856i,used=185856i,used_percent=100 1453832006274137913
disk,fstype=autofs,mode=rw,path=/net free=0i,inodes_free=0i,inodes_total=0i,inodes_used=0i,total=0i,used=0i,used_percent=0 1453832006274157077
disk,fstype=autofs,mode=rw,path=/home free=0i,inodes_free=0i,inodes_total=0i,inodes_used=0i,total=0i,used=0i,used_percent=0 1453832006274169688
disk,device=dm-1,fstype=xfs,label=lvg-lv,mode=rw,path=/mnt inodes_free=8388605i,inodes_used=3i,total=17112760320i,free=16959598592i,used=153161728i,used_percent=0.8950147441789215,inodes_total=8388608i 1677001387000000000
```

[statfs]: http://man7.org/linux/man-pages/man2/statfs.2.html
