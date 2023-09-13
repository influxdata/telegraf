# Lustre Input Plugin

The [Lustre][]® file system is an open-source, parallel file system that
supports many requirements of leadership class HPC simulation environments.

This plugin monitors the Lustre file system using its utility `lctl get_param`,
which is the standard and recommanded way to monitor and 
statistics[Lustre Monitoring and Statistics Guide][guide].

Note that this plugins has been only tested on Lustre@v2.12.7 
and Luster@v2.15.0.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about Lustre components, including ost/oss, mdt/mds, client.
# This plugin ONLY supports Linux.
[[inputs.lustre2_lctl]]
  ## According to different components, you could choose to gather pointed data about the component.

  ost_collect  = [
    "obdfilter.*.stats",
    "obdfilter.*.job_stats",
    "obdfilter.*.recovery_status",
    "obdfilter.*.kbytesfree", # osd-ldiskfs.*.kbytesfree, osd-zfs.*.kbytesfree
    "obdfilter.*.kbytesavail", # osd-ldiskfs.*.kbytesavail, osd-zfs.*.kbytesavail
    "obdfilter.*.kbytestotal", # osd-ldiskfs.*.kbytestotal, osd-zfs.*.kbytestotal
  ]

  mdt_collect = [
    "mdt.*.recovery_status",
    "mdt.*.stats",
    "mdt.*.job_stats",
  ]

  client_collect = [
    "osc.*.active",
    "mdc.*.active",
  ]

```

## Metrics

### OST

* tags
  * volume (the name of volume)
  * jobid
  * unit
* fields
  * ost_health_check (uint)
  * ost_recovery_status (uint)
  * ost_jobstats_*_samples (uint)
  * ost_jobstats_*_max     (uint)
  * ost_jobstats_*_min     (uint)
  * ost_jobstats_*_sum     (uint)
  * ost_jobstats_*_sumsq   (uint)
  * ost_capacity_kbytestotal (uint)
  * ost_capacity_kbytesavail (uint)
  * ost_capacity_kbytesfree (uint)

### MDT

* tags
  * volume (the name of volume)
  * jobid
  * unit
* fields
  * mdt_health_check (uint)
  * mdt_recovery_status (uint)
  * mdt_jobstats_*_samples (uint)
  * mdt_jobstats_*_max     (uint)
  * mdt_jobstats_*_min     (uint)
  * mdt_jobstats_*_sum     (uint)
  * mdt_jobstats_*_sumsq   (uint)

### Client

* tags
  * volume
* fields
  * osc_volume_active (uint)
  * mdc_volume_active (uint)

## Troubleshooting

Check for the default or custom procfiles in the proc filesystem, and reference
the [Lustre Monitoring and Statistics Guide][guide].  This plugin does not
report all information from these files, only a limited set of items
corresponding to the above metric fields.

## Example Output

```text
lustre2_client_osc_volume_active{host="ln0",volume="THL9-OST001f"} 1
lustre2_client_mdc_volume_active{host="ln0",volume="THL9-MDT0000"} 1
lustre2_client_health_check{host="ln0"} 1
```

[lustre]: http://lustre.org/
[guide]: http://wiki.lustre.org/Lustre_Monitoring_and_Statistics_Guide
