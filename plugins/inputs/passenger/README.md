# Passenger Input Plugin

This plugin gathers metrics from the [Phusion Passenger][phusion] service.

> [!WARNING]
> Depending on your environment, this plugin can create a high number of series
> which can cause high load on your database. Please use
> [measurement filtering][metric_filtering] to manage your series cardinality!

The plugin uses the `passenger-status` command line tool.

> [!NOTE]
> This plugin requires the `passenger-status` binary to be installed on the
> system and to be executable by Telegraf.

⭐ Telegraf v0.10.1
🏷️ web
💻 all

[phusion]: https://www.phusionpassenger.com/
[metric_filtering]: /docs/CONFIGURATION.md#metric-filtering

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics of passenger using passenger-status
[[inputs.passenger]]
  ## Path of passenger-status.
  ##
  ## Plugin gather metric via parsing XML output of passenger-status
  ## More information about the tool:
  ##   https://www.phusionpassenger.com/library/admin/apache/overall_status_report.html
  ##
  ## If no path is specified, then the plugin simply execute passenger-status
  ## hopefully it can be found in your PATH
  command = "passenger-status -v --show=xml"
```

### Permissions

Telegraf must have permission to execute the `passenger-status` command.  On
most systems, Telegraf runs as the `telegraf` user.

### Series Cardinality

Depending on your environment, this `passenger_process` measurement of this
plugin can quickly create a high number of series which, when unchecked, can
cause high load on your database.  You can use the following techniques to
manage your series cardinality:

- Use the
  [measurement filtering](https://docs.influxdata.com/telegraf/latest/administration/configuration/#measurement-filtering)
  options to exclude unneeded tags.  In some environments, you may wish to use
  `tagexclude` to remove the `pid` and `process_group_id` tags.
- Write to a database with an appropriate
  [retention policy](https://docs.influxdata.com/influxdb/latest/guides/downsampling_and_retention/).
- Consider using the
  [Time Series Index](https://docs.influxdata.com/influxdb/latest/concepts/time-series-index/).
- Monitor your databases
  [series cardinality](https://docs.influxdata.com/influxdb/latest/query_language/spec/#show-cardinality).

## Metrics

- passenger
  - tags:
    - passenger_version
  - fields:
    - process_count
    - max
    - capacity_used
    - get_wait_list_size

- passenger_supergroup
  - tags:
    - name
  - fields:
    - get_wait_list_size
    - capacity_used

- passenger_group
  - tags:
    - name
    - app_root
    - app_type
  - fields:
    - get_wait_list_size
    - capacity_used
    - processes_being_spawned

- passenger_process
  - tags:
    - group_name
    - app_root
    - supergroup_name
    - pid
    - code_revision
    - life_status
    - process_group_id
  - fields:
    - concurrency
    - sessions
    - busyness
    - processed
    - spawner_creation_time
    - spawn_start_time
    - spawn_end_time
    - last_used
    - uptime
    - cpu
    - rss
    - pss
    - private_dirty
    - swap
    - real_memory
    - vmsize

## Example Output

```text
passenger,passenger_version=5.0.17 capacity_used=23i,get_wait_list_size=0i,max=23i,process_count=23i 1452984112799414257
passenger_supergroup,name=/var/app/current/public capacity_used=23i,get_wait_list_size=0i 1452984112799496977
passenger_group,app_root=/var/app/current,app_type=rack,name=/var/app/current/public capacity_used=23i,get_wait_list_size=0i,processes_being_spawned=0i 1452984112799527021
passenger_process,app_root=/var/app/current,code_revision=899ac7f,group_name=/var/app/current/public,life_status=ALIVE,pid=11553,process_group_id=13608,supergroup_name=/var/app/current/public busyness=0i,concurrency=1i,cpu=58i,last_used=1452747071764940i,private_dirty=314900i,processed=951i,pss=319391i,real_memory=314900i,rss=418548i,sessions=0i,spawn_end_time=1452746845013365i,spawn_start_time=1452746844946982i,spawner_creation_time=1452746835922747i,swap=0i,uptime=226i,vmsize=1563580i 1452984112799571490
passenger_process,app_root=/var/app/current,code_revision=899ac7f,group_name=/var/app/current/public,life_status=ALIVE,pid=11563,process_group_id=13608,supergroup_name=/var/app/current/public busyness=2147483647i,concurrency=1i,cpu=47i,last_used=1452747071709179i,private_dirty=309240i,processed=756i,pss=314036i,real_memory=309240i,rss=418296i,sessions=1i,spawn_end_time=1452746845172460i,spawn_start_time=1452746845136882i,spawner_creation_time=1452746835922747i,swap=0i,uptime=226i,vmsize=1563608i 1452984112799638581
```
