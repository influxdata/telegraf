# Lustre Input Plugin

This plugin gathers metrics for the [Lustre® file system][lustre] using its
entries in the `proc` filesystem. Reference the
[Lustre Monitoring and Statistics Guide][guide] for the reported information.

> [!NOTE] This plugin doesn't report _all_ information available but only a
> limited set of items. Check the [metrics section](#metrics).

⭐ Telegraf v0.1.5
🏷️ system
💻 linux

[lustre]: http://lustre.org/
[guide]: http://wiki.lustre.org/Lustre_Monitoring_and_Statistics_Guide

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from local Lustre service on OST, MDS
# This plugin ONLY supports Linux
[[inputs.lustre2]]
  ## An array of /proc globs to search for Lustre stats
  ## If not specified, the default will work on Lustre 2.12.x and newer
  ## versions like 2.17.x, which moved many files to /sys/kernel/debug/lustre
  ## (usually only readable by root)
  ##
  # mgs_procfiles = [
  #   "/sys/fs/lustre/mgs/*/eviction_count",
  # ]
  # ost_procfiles = [
  #   "/proc/fs/lustre/obdfilter/*/stats",
  #   "/proc/fs/lustre/osd-ldiskfs/*/stats",
  #   "/proc/fs/lustre/obdfilter/*/job_stats",
  #   "/proc/fs/lustre/obdfilter/*/exports/*/stats",
  #   "/proc/fs/lustre/osd-ldiskfs/*/brw_stats",
  #   "/proc/fs/lustre/osd-zfs/*/brw_stats",
  #   "/sys/fs/lustre/odbfilter/*/eviction_count",
  #   "/sys/kernel/debug/lustre/obdfilter/*/stats",
  #   "/sys/kernel/debug/lustre/osd-ldiskfs/*/stats",
  #   "/sys/kernel/debug/lustre/osd-zfs/*/stats",
  #   "/sys/kernel/debug/lustre/osd-ldiskfs/*/brw_stats",
  #   "/sys/kernel/debug/lustre/osd-zfs/*/brw_stats",
  # ]
  # mds_procfiles = [
  #   "/proc/fs/lustre/mdt/*/md_stats",
  #   "/proc/fs/lustre/mdt/*/job_stats",
  #   "/proc/fs/lustre/mdt/*/exports/*/stats",
  #   "/proc/fs/lustre/osd-ldiskfs/*/brw_stats",
  #   "/proc/fs/lustre/osd-zfs/*/brw_stats",
  #   "/sys/fs/lustre/mdt/*/eviction_count",
  # ]
```

Newer Lustre versions (e.g. 2.17) moved many statistics files to
`/sys/kernel/debug/lustre`, e.g. `obdfilter/*/stats`, `osd-*/*/stats`,
`osd-*/*/brw_stats`, `*/*/exports/*/stats` and the RPC service statistics in
`ost/OSS/*/stats` or `mds/MDS/*/stats`. The defaults include the new
locations of the `obdfilter` and `osd-*` statistics, all others have to be
configured explicitly. Note that `/sys/kernel/debug` is usually only accessible
by `root`, so Telegraf must be run with sufficient permissions to read them.

## Metrics

From `/sys/fs/lustre/health_check`:

- lustre2
  - tags:
  - fields:
    - health

From `/proc/fs/lustre/obdfilter/*/stats` and
`/proc/fs/lustre/osd-ldiskfs/*/stats`:

- lustre2
  - tags:
    - name
  - fields:
    - write_bytes
    - write_calls
    - read_bytes
    - read_calls
    - cache_hit (samples, retained for backward compatibility)
    - cache_hit_samples
    - cache_hit_pages
    - cache_miss (samples, retained for backward compatibility)
    - cache_miss_samples
    - cache_miss_pages
    - cache_access (samples, retained for backward compatibility)
    - cache_access_samples
    - cache_access_pages
    - copy_pages (`osd-zfs` only)
    - zerocopy_pages (`osd-zfs` only)
    - tail_pages (`osd-zfs` only)

From `/proc/fs/lustre/osd-ldiskfs/*/brw_stats` and
`/proc/fs/lustre/osd-zfs/*/brw_stats`:

- lustre2
  - tags:
    - name
    - brw_section (one of `pages_per_bulk_rw`, `discontiguous_pages`,
      `discontiguous_blocks`, `disk_fragmented_ios`, `disk_ios_in_flight`,
      `io_time`, `disk_io_size` and `block_maps_msec`)
    - bucket
  - fields:
    - read_ios
    - read_percent
    - write_ios
    - write_percent

From `/proc/fs/lustre/obdfilter/*/exports/*/stats`:

- lustre2
  - tags:
    - name
    - client
  - fields:
    - write_bytes
    - write_calls
    - read_bytes
    - read_calls

From `/proc/fs/lustre/obdfilter/*/job_stats`:

- lustre2
  - tags:
    - name
    - jobid
  - fields:
    - jobstats_ost_getattr
    - jobstats_ost_setattr
    - jobstats_ost_sync
    - jobstats_punch
    - jobstats_destroy
    - jobstats_create
    - jobstats_ost_statfs
    - jobstats_get_info
    - jobstats_set_info
    - jobstats_quotactl
    - jobstats_read_bytes
    - jobstats_read_calls
    - jobstats_read_max_size
    - jobstats_read_min_size
    - jobstats_write_bytes
    - jobstats_write_calls
    - jobstats_write_max_size
    - jobstats_write_min_size
    - jobstats_read_time (sum of read latencies in microseconds, newer versions)
    - jobstats_read_min_time (in microseconds, newer versions)
    - jobstats_read_max_time (in microseconds, newer versions)
    - jobstats_write_time (sum of write latencies in microseconds, newer
      versions)
    - jobstats_write_min_time (in microseconds, newer versions)
    - jobstats_write_max_time (in microseconds, newer versions)

From `/proc/fs/lustre/mdt/*/md_stats`:

- lustre2
  - tags:
    - name
  - fields:
    - open
    - close
    - mknod
    - link
    - unlink
    - mkdir
    - rmdir
    - rename
    - getattr
    - setattr
    - getxattr
    - setxattr
    - statfs
    - sync
    - samedir_rename
    - crossdir_rename
    - parallel_rename_file
    - parallel_rename_dir
    - rename_trylocks
    - punch
    - read_bytes
    - read_calls
    - write_bytes
    - write_calls

From `/proc/fs/lustre/mdt/*/exports/*/stats`:

- lustre2
  - tags:
    - name
    - client
  - fields:
    - open
    - close
    - mknod
    - link
    - unlink
    - mkdir
    - rmdir
    - rename
    - getattr
    - setattr
    - getxattr
    - setxattr
    - statfs
    - sync
    - samedir_rename
    - crossdir_rename
    - parallel_rename_file
    - parallel_rename_dir
    - rename_trylocks
    - punch
    - read_bytes
    - read_calls
    - write_bytes
    - write_calls

From `/proc/fs/lustre/mdt/*/job_stats`:

- lustre2
  - tags:
    - name
    - jobid
  - fields:
    - jobstats_close
    - jobstats_crossdir_rename
    - jobstats_getattr
    - jobstats_getxattr
    - jobstats_link
    - jobstats_mkdir
    - jobstats_mknod
    - jobstats_open
    - jobstats_rename
    - jobstats_rmdir
    - jobstats_samedir_rename
    - jobstats_setattr
    - jobstats_setxattr
    - jobstats_statfs
    - jobstats_sync
    - jobstats_unlink
    - jobstats_parallel_rename_file
    - jobstats_parallel_rename_dir
    - jobstats_rename_trylocks
    - jobstats_migrate
    - jobstats_fallocate
    - jobstats_punch
    - jobstats_read_bytes
    - jobstats_read_calls
    - jobstats_read_max_size
    - jobstats_read_min_size
    - jobstats_write_bytes
    - jobstats_write_calls
    - jobstats_write_max_size
    - jobstats_write_min_size
    - jobstats_read_time (sum of read latencies in microseconds)
    - jobstats_read_min_time (in microseconds)
    - jobstats_read_max_time (in microseconds)
    - jobstats_write_time (sum of write latencies in microseconds)
    - jobstats_write_min_time (in microseconds)
    - jobstats_write_max_time (in microseconds)

From RPC statistics files like `ost/OSS/*/stats`, `mds/MDS/*/stats`,
`ldlm/services/*/stats`, `lwp/*/stats` or `osp/*/stats`, i.e. files containing
a `req_waittime` counter, all counters are reported. The `name` tag contains the
service or import, e.g. `ost_io` or `fs-OST0000-osc-MDT0000`.

- lustre2
  - tags:
    - name
  - fields:
    - `<counter>_samples` (number of samples, e.g. `ost_read_samples`)
    - `<counter>_sum` (sum of the sampled values in the unit stated in the file,
      e.g. `ost_read_sum` in microseconds)

From `/proc/fs/lustre/*/*/eviction_count`:

- lustre2
  - tags:
    - name
  - fields:
    - evictions

## Example Output

```text
lustre2,host=oss2,jobid=42990218,name=wrk-OST0041 jobstats_ost_setattr=0i,jobstats_ost_sync=0i,jobstats_punch=0i,jobstats_read_bytes=4096i,jobstats_read_calls=1i,jobstats_read_max_size=4096i,jobstats_read_min_size=4096i,jobstats_write_bytes=310206488i,jobstats_write_calls=7423i,jobstats_write_max_size=53048i,jobstats_write_min_size=8820i 1556525847000000000
lustre2,host=mds1,jobid=42992017,name=wrk-MDT0000 jobstats_close=31798i,jobstats_crossdir_rename=0i,jobstats_getattr=34146i,jobstats_getxattr=15i,jobstats_link=0i,jobstats_mkdir=658i,jobstats_mknod=0i,jobstats_open=31797i,jobstats_rename=0i,jobstats_rmdir=0i,jobstats_samedir_rename=0i,jobstats_setattr=1788i,jobstats_setxattr=0i,jobstats_statfs=0i,jobstats_sync=0i,jobstats_unlink=0i 1556525828000000000
```
