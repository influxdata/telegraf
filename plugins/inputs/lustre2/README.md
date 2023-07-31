# Lustre Input Plugin 
The [Lustre][]® file system is an open-source, parallel file system that
supports many requirements of leadership class HPC simulation environments.

This plugin monitors the Lustre file system using its entries in the proc
filesystem. Compared to https://github.com/influxdata/telegraf/tree/master/plugins/inputs/lustre2,
we use commond `lctl get_param` to get statistics instead of reading data based on absolute path of statistics file(recommand from the Lustre).

## Global configuration options <!-- @/docs/includes/plugin_config.md -->
In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration
```toml @sample.conf
# Read metrics from local Lustre service on OST, MDS
# This plugin ONLY supports Linux
[[inputs.lustre2]]
  [[inputs.lustre2]]
  ## A variable indicates the identifier of a node, because different types of nodes
  ## have differenct metrics.
  ##
  # ost = true | false
  # mdt = true | false
  # mgs = true | false
  
```

## Metrics

|                     name                     | type  |                               value                                |       description       |
| :------------------------------------------: | :---: | :----------------------------------------------------------------: | :---------------------: |
|           lustre2_ost_health_check           | Gauge |             1 represent healthy, 0 represent unhealthy             |            -            |
|         lustre2_ost_recovery_status          | Gauge | 1 represents recovery completed, 0 represents recovery uncompleted |            -            |
|   lustre2_ost_jobstats_read_bytes_samples    | Gauge |               the number of read operations of a job               |            -            |
|     lustre2_ost_jobstats_read_bytes_min      | Gauge |           the maximum bytes of a read operation of a job           |            -            |
|     lustre2_ost_jobstats_read_bytes_max      | Gauge |           the minimum bytes of a read operation of a job           |            -            |
|     lustre2_ost_jobstats_read_bytes_sum      | Gauge |            the total bytes of read operations of a job             |            -            |
|    lustre2_ost_jobstats_read_bytes_sumsq     | Gauge |                                 -                                  | support >= lustre v2.15 |
|   lustre2_ost_jobstats_write_bytes_samples   | Gauge |              the number of write operations of a job               |            -            |
|     lustre2_ost_jobstats_write_bytes_min     | Gauge |          the maximum bytes of a write operation of a job           |            -            |
|     lustre2_ost_jobstats_write_bytes_max     | Gauge |          the minimum bytes of a write operation of a job           |            -            |
|     lustre2_ost_jobstats_write_bytes_sum     | Gauge |            the total bytes of write operations of a job            |            -            |
|    lustre2_ost_jobstats_write_bytes_sumsq    | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_getattr_samples     | Gauge |                                 -                                  |            -            |
|       lustre2_ost_jobstats_getattr_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_getattr_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_getattr_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_getattr_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_setattr_samples     | Gauge |                                 -                                  |            -            |
|       lustre2_ost_jobstats_setattr_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_setattr_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_setattr_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_setattr_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_punch_samples      | Gauge |                                 -                                  |            -            |
|        lustre2_ost_jobstats_punch_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_ost_jobstats_punch_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_ost_jobstats_punch_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_punch_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_sync_samples       | Gauge |                                 -                                  |            -            |
|        lustre2_ost_jobstats_sync_min         | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_ost_jobstats_sync_max         | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_ost_jobstats_sync_sum         | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_sync_sumsq        | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_destroy_samples     | Gauge |                                 -                                  |            -            |
|       lustre2_ost_jobstats_destroy_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_destroy_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_destroy_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_destroy_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_create_samples      | Gauge |                                 -                                  |            -            |
|       lustre2_ost_jobstats_create_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_create_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_create_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_create_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_statfs_samples      | Gauge |                                 -                                  |            -            |
|       lustre2_ost_jobstats_statfs_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_statfs_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_ost_jobstats_statfs_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_statfs_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_ost_jobstats_get_info_samples     | Gauge |                                 -                                  |            -            |
|      lustre2_ost_jobstats_get_info_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_get_info_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_get_info_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_get_info_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_ost_jobstats_set_info_samples     | Gauge |                                 -                                  |            -            |
|      lustre2_ost_jobstats_set_info_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_set_info_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_set_info_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_set_info_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_ost_jobstats_quotactl_samples     | Gauge |                                 -                                  |            -            |
|      lustre2_ost_jobstats_quotactl_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_quotactl_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_quotactl_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_quotactl_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_ost_jobstats_prealloc_samples     | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_prealloc_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_prealloc_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_ost_jobstats_prealloc_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_ost_jobstats_prealloc_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_open_samples       | Gauge |                                 -                                  |            -            |
|        lustre2_mdt_jobstats_open_min         | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_open_max         | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_open_sum         | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_open_sumsq        | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_close_samples      | Gauge |                                 -                                  |            -            |
|        lustre2_mdt_jobstats_close_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_close_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_close_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_close_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_mknod_samples      | Gauge |                                 -                                  |            -            |
|        lustre2_mdt_jobstats_mknod_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_mknod_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_mknod_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_mknod_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_link_samples       | Gauge |                                 -                                  |            -            |
|        lustre2_mdt_jobstats_link_min         | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_link_max         | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_link_sum         | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_link_sumsq        | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_unlink_samples      | Gauge |                                 -                                  |            -            |
|       lustre2_mdt_jobstats_unlink_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_unlink_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_unlink_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_unlink_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_mkdir_samples      | Gauge |                                 -                                  |            -            |
|        lustre2_mdt_jobstats_mkdir_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_mkdir_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_mkdir_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_mkdir_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_rmdir_samples      | Gauge |                                 -                                  |            -            |
|        lustre2_mdt_jobstats_rmdir_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_rmdir_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_rmdir_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_rmdir_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_rename_samples      | Gauge |                                 -                                  |            -            |
|       lustre2_mdt_jobstats_rename_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_rename_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_rename_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_rename_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_getattr_samples     | Gauge |                                 -                                  |            -            |
|       lustre2_mdt_jobstats_getattr_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_getattr_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_getattr_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_getattr_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_setattr_samples     | Gauge |                                 -                                  |            -            |
|       lustre2_mdt_jobstats_setattr_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_setattr_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_setattr_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_setattr_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_mdt_jobstats_getxattr_samples     | Gauge |                                 -                                  |            -            |
|      lustre2_mdt_jobstats_getxattr_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_getxattr_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_getxattr_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_getxattr_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_mdt_jobstats_setxattr_samples     | Gauge |                                 -                                  |            -            |
|      lustre2_mdt_jobstats_setxattr_min       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_setxattr_max       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_setxattr_sum       | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_setxattr_sumsq      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_statfs_samples      | Gauge |                                 -                                  |            -            |
|       lustre2_mdt_jobstats_statfs_min        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_statfs_max        | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_statfs_sum        | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_statfs_sumsq       | Gauge |                                 -                                  | support >= lustre v2.15 |
|      lustre2_mdt_jobstats_sync_samples       | Gauge |                                 -                                  |            -            |
|        lustre2_mdt_jobstats_sync_min         | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_sync_max         | Gauge |                                 -                                  | support >= lustre v2.15 |
|        lustre2_mdt_jobstats_sync_sum         | Gauge |                                 -                                  | support >= lustre v2.15 |
|       lustre2_mdt_jobstats_sync_sumsq        | Gauge |                                 -                                  | support >= lustre v2.15 |
| lustre2_mdt_jobstats_samedir_rename_samples  | Gauge |                                 -                                  |            -            |
|   lustre2_mdt_jobstats_samedir_rename_min    | Gauge |                                 -                                  | support >= lustre v2.15 |
|   lustre2_mdt_jobstats_samedir_rename_max    | Gauge |                                 -                                  | support >= lustre v2.15 |
|   lustre2_mdt_jobstats_samedir_rename_sum    | Gauge |                                 -                                  | support >= lustre v2.15 |
|  lustre2_mdt_jobstats_samedir_rename_sumsq   | Gauge |                                 -                                  | support >= lustre v2.15 |
| lustre2_mdt_jobstats_crossdir_rename_samples | Gauge |                                 -                                  |            -            |
|   lustre2_mdt_jobstats_crossdir_rename_min   | Gauge |                                 -                                  | support >= lustre v2.15 |
|   lustre2_mdt_jobstats_crossdir_rename_max   | Gauge |                                 -                                  | support >= lustre v2.15 |
|   lustre2_mdt_jobstats_crossdir_rename_sum   | Gauge |                                 -                                  | support >= lustre v2.15 |
|  lustre2_mdt_jobstats_crossdir_rename_sumsq  | Gauge |                                 -                                  | support >= lustre v2.15 |
|   lustre2_mdt_jobstats_read_bytes_samples    | Gauge |                                 -                                  |            -            |
|     lustre2_mdt_jobstats_read_bytes_min      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_read_bytes_max      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_read_bytes_sum      | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_mdt_jobstats_read_bytes_sumsq     | Gauge |                                 -                                  | support >= lustre v2.15 |
|   lustre2_mdt_jobstats_write_bytes_samples    | Gauge |                                 -                                  |            -            |
|     lustre2_mdt_jobstats_write_bytes_min      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_write_bytes_max      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_write_bytes_sum      | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_mdt_jobstats_write_bytes_sumsq     | Gauge |                                 -                                  | support >= lustre v2.15 |
|   lustre2_mdt_jobstats_punch_bytes_samples    | Gauge |                                 -                                  |            -            |
|     lustre2_mdt_jobstats_punch_bytes_min      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_punch_bytes_max      | Gauge |                                 -                                  | support >= lustre v2.15 |
|     lustre2_mdt_jobstats_punch_bytes_sum      | Gauge |                                 -                                  | support >= lustre v2.15 |
|    lustre2_mdt_jobstats_punch_bytes_sumsq     | Gauge |                                 -                                  | support >= lustre v2.15 |

## Example Output

```text
lustre2_ost_ost_jobstats_create_samples{cluster="hpc4",host="ost114",jobid="1211707",unit="reqs",volume="THL9-OST0005"} 0
lustre2_ost_ost_jobstats_create_samples{cluster="hpc4",host="ost114",jobid="1228445",unit="reqs",volume="THL9-OST0004"} 0
lustre2_ost_ost_jobstats_create_samples{cluster="hpc4",host="ost114",jobid="1230486",unit="reqs",volume="THL9-OST0005"} 0
lustre2_ost_ost_jobstats_create_samples{cluster="hpc4",host="ost114",jobid="1233543",unit="reqs",volume="THL9-OST0005"} 0
lustre2_ost_ost_jobstats_create_samples{cluster="hpc4",host="ost114",jobid="1235122",unit="reqs",volume="THL9-OST0004"} 0
```

[lustre]: http://lustre.org/
[guide]: http://wiki.lustre.org/Lustre_Monitoring_and_Statistics_Guide