# Mesos Input Plugin

This input plugin gathers metrics from Mesos (*currently only Mesos masters*).
For more information, please check the [Mesos Observability Metrics](http://mesos.apache.org/documentation/latest/monitoring/) page.

### Configuration:

```toml
# Telegraf plugin for gathering metrics from N Mesos masters
[[inputs.mesos]]
  # Timeout, in ms.
  timeout = 100
  # A list of Mesos masters, default value is localhost:5050.
  masters = ["localhost:5050"]
  # Metrics groups to be collected, by default, all enabled.
  master_collections = ["resources","master","system","slaves","frameworks","messages","evqueue","registrar"]
```

### Measurements & Fields:

Mesos master metric groups

- resources
    - master/cpus_percent
    - master/cpus_used
    - master/cpus_total
    - master/cpus_revocable_percent
    - master/cpus_revocable_total
    - master/cpus_revocable_used
    - master/disk_percent
    - master/disk_used
    - master/disk_total
    - master/disk_revocable_percent
    - master/disk_revocable_total
    - master/disk_revocable_used
    - master/mem_percent
    - master/mem_used
    - master/mem_total
    - master/mem_revocable_percent
    - master/mem_revocable_total
    - master/mem_revocable_used

- master
    - master/elected
    - master/uptime_secs

- system
    - system/cpus_total
    - system/load_15min
    - system/load_5min
    - system/load_1min
    - system/mem_free_bytes
    - system/mem_total_bytes

- slaves
    - master/slave_registrations
    - master/slave_removals
    - master/slave_reregistrations
    - master/slave_shutdowns_scheduled
    - master/slave_shutdowns_canceled
    - master/slave_shutdowns_completed
    - master/slaves_active
    - master/slaves_connected
    - master/slaves_disconnected
    - master/slaves_inactive

- frameworks
    - master/frameworks_active
    - master/frameworks_connected
    - master/frameworks_disconnected
    - master/frameworks_inactive
    - master/outstanding_offers

- tasks
    - master/tasks_error
    - master/tasks_failed
    - master/tasks_finished
    - master/tasks_killed
    - master/tasks_lost
    - master/tasks_running
    - master/tasks_staging
    - master/tasks_starting

- messages
    - master/invalid_executor_to_framework_messages
    - master/invalid_framework_to_executor_messages
    - master/invalid_status_update_acknowledgements
    - master/invalid_status_updates
    - master/dropped_messages
    - master/messages_authenticate
    - master/messages_deactivate_framework
    - master/messages_decline_offers
    - master/messages_executor_to_framework
    - master/messages_exited_executor
    - master/messages_framework_to_executor
    - master/messages_kill_task
    - master/messages_launch_tasks
    - master/messages_reconcile_tasks
    - master/messages_register_framework
    - master/messages_register_slave
    - master/messages_reregister_framework
    - master/messages_reregister_slave
    - master/messages_resource_request
    - master/messages_revive_offers
    - master/messages_status_update
    - master/messages_status_update_acknowledgement
    - master/messages_unregister_framework
    - master/messages_unregister_slave
    - master/messages_update_slave
    - master/recovery_slave_removals
    - master/slave_removals/reason_registered
    - master/slave_removals/reason_unhealthy
    - master/slave_removals/reason_unregistered
    - master/valid_framework_to_executor_messages
    - master/valid_status_update_acknowledgements
    - master/valid_status_updates
    - master/task_lost/source_master/reason_invalid_offers
    - master/task_lost/source_master/reason_slave_removed
    - master/task_lost/source_slave/reason_executor_terminated
    - master/valid_executor_to_framework_messages

- evqueue
    - master/event_queue_dispatches
    - master/event_queue_http_requests
    - master/event_queue_messages

- registrar
    - registrar/state_fetch_ms
    - registrar/state_store_ms
    - registrar/state_store_ms/max
    - registrar/state_store_ms/min
    - registrar/state_store_ms/p50
    - registrar/state_store_ms/p90
    - registrar/state_store_ms/p95
    - registrar/state_store_ms/p99
    - registrar/state_store_ms/p999
    - registrar/state_store_ms/p9999

### Tags:

- All measurements have the following tags:
    - server

### Example Output:

```
$ telegraf -config ~/mesos.conf -input-filter mesos -test
* Plugin: mesos, Collection 1
mesos,server=172.17.8.101 allocator/event_queue_dispatches=0,master/cpus_percent=0,
master/cpus_revocable_percent=0,master/cpus_revocable_total=0,
master/cpus_revocable_used=0,master/cpus_total=2,
master/cpus_used=0,master/disk_percent=0,master/disk_revocable_percent=0,
master/disk_revocable_total=0,master/disk_revocable_used=0,master/disk_total=10823,
master/disk_used=0,master/dropped_messages=2,master/elected=1,
master/event_queue_dispatches=10,master/event_queue_http_requests=0,
master/event_queue_messages=0,master/frameworks_active=2,master/frameworks_connected=2,
master/frameworks_disconnected=0,master/frameworks_inactive=0,
master/invalid_executor_to_framework_messages=0,
master/invalid_framework_to_executor_messages=0,
master/invalid_status_update_acknowledgements=0,master/invalid_status_updates=0,master/mem_percent=0,
master/mem_revocable_percent=0,master/mem_revocable_total=0,
master/mem_revocable_used=0,master/mem_total=1002,
master/mem_used=0,master/messages_authenticate=0,
master/messages_deactivate_framework=0 ...
```
