# Puppet Agent Input Plugin

This plugin gathers metrics of a [Puppet agent][puppet] by parsing variables
from the local last-run-summary file.

⭐ Telegraf v0.2.0
🏷️ system
💻 all

[puppet]: https://www.puppet.com/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Reads last_run_summary.yaml file and converts to measurements
[[inputs.puppetagent]]
  ## Location of puppet last run summary file
  location = "/var/lib/puppet/state/last_run_summary.yaml"
```

## Metrics

### PuppetAgent int64 measurements

Meta:

- units: int64
- tags: ``

Measurement names:

- puppetagent_changes_total
- puppetagent_events_failure
- puppetagent_events_total
- puppetagent_events_success
- puppetagent_resources_changed
- puppetagent_resources_corrective_change
- puppetagent_resources_failed
- puppetagent_resources_failedtorestart
- puppetagent_resources_outofsync
- puppetagent_resources_restarted
- puppetagent_resources_scheduled
- puppetagent_resources_skipped
- puppetagent_resources_total
- puppetagent_time_service
- puppetagent_time_lastrun
- puppetagent_version_config

### PuppetAgent float64 measurements

Meta:

- units: float64
- tags: ``

Measurement names:

- puppetagent_time_anchor
- puppetagent_time_catalogapplication
- puppetagent_time_configretrieval
- puppetagent_time_convertcatalog
- puppetagent_time_cron
- puppetagent_time_exec
- puppetagent_time_factgeneration
- puppetagent_time_file
- puppetagent_time_filebucket
- puppetagent_time_group
- puppetagent_time_lastrun
- puppetagent_time_noderetrieval
- puppetagent_time_notify
- puppetagent_time_package
- puppetagent_time_pluginsync
- puppetagent_time_schedule
- puppetagent_time_sshauthorizedkey
- puppetagent_time_total
- puppetagent_time_transactionevaluation
- puppetagent_time_user
- puppetagent_version_config

### PuppetAgent string measurements

Meta:

- units: string
- tags: ``

Measurement names:

- puppetagent_version_puppet

## Example Output

```text
puppetagent,location=last_run_summary.yaml changes_total=0i,events_failure=0i,events_noop=0i,events_success=0i,events_total=0i,resources_changed=0i,resources_correctivechange=0i,resources_failed=0i,resources_failedtorestart=0i,resources_outofsync=0i,resources_restarted=0i,resources_scheduled=0i,resources_skipped=0i,resources_total=109i,time_anchor=0.000555,time_catalogapplication=0.010555,time_configretrieval=4.75567007064819,time_convertcatalog=1.3,time_cron=0.000584,time_exec=0.508123,time_factgeneration=0.34,time_file=0.441472,time_filebucket=0.000353,time_group=0,time_lastrun=1444936531i,time_noderetrieval=1.235,time_notify=0.00035,time_package=1.325788,time_pluginsync=0.325788,time_schedule=0.001123,time_service=1.807795,time_sshauthorizedkey=0.000764,time_total=8.85354707064819,time_transactionevaluation=4.69765,time_user=0.004331,version_configstring="environment:d6018ce",version_puppet="3.7.5" 1747757240432097335
```
