# Lustre lctl Input Plugin

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

* lustre2_lctl
  * tags
    * volume (the name of volume)
    * jobid
    * unit
  * fields
    * health_check (int)
    * ost_recovery_status (int)
    * ost_jobstats_*_samples (int)
    * ost_jobstats_*_max     (int)
    * ost_jobstats_*_min     (int)
    * ost_jobstats_*_sum     (int)
    * ost_jobstats_*_sumsq   (int)
    * ost_stats_*_samples (int)
    * ost_stats_*_min (int)
    * ost_stats_*_max (int)
    * ost_stats_*_sum (int)
    * ost_stats_*_sumsq (int)
    * ost_capacity_kbytestotal (int)
    * ost_capacity_kbytesavail (int)
    * ost_capacity_kbytesfree (int)

### MDT

* lustre2_lclt
  * tags
    * volume (the name of volume)
    * jobid
    * unit
  * fields
    * mdt_recovery_status (int)
    * mdt_jobstats_*_samples (int)
    * mdt_jobstats_*_max     (int)
    * mdt_jobstats_*_min     (int)
    * mdt_jobstats_*_sum     (int)
    * mdt_jobstats_*_sumsq   (int)
    * mdt_stats_*_sample (int)
    * mdt_stats_*_min (int)
    * mdt_stats_*_max (int)
    * mdt_stats_*_sum (int)
    * mdt_stats_*_sumsq (int)

### Client

* lustre2_lctl
  * tags
    * volume
  * fields
    * osc_volume_active (int)
    * mdc_volume_active (int)

## Troubleshooting

Check for the default or custom procfiles in the proc filesystem, and reference
the [Lustre Monitoring and Statistics Guide][guide].  This plugin does not
report all information from these files, only a limited set of items
corresponding to the above metric fields.

## Example Output

```text
lustre2_lctl,host=ost114 health_check=1i 1695018430000000000
lustre2_lctl_ost,host=ost114,volume=OST0004 recovery_status=1i 1695018430000000000
lustre2_lctl_ost,host=ost114,volume=OST0005 recovery_status=1i 1695018430000000000
lustre2_lctl_ost,host=ost114,volume=OST0004 capacity_kbytestotal=46488188776i 1695018430000000000
```

[lustre]: http://lustre.org/
[guide]: http://wiki.lustre.org/Lustre_Monitoring_and_Statistics_Guide
