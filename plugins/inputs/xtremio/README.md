# XtremIO Input Plugin

The `xtremio` plugin gathers metrics from a Dell EMC XtremIO Storage Array's V3 Rest API.
### Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage xtremio`.

```toml
[[inputs.xtremio]]
  ## XtremIO Username
  username = "user1" # required
  ## XtremIO Password
  password = "pass123" # required
  ## XtremIO User Interface Endpoint
  url = "https://xtremio.example.com/" # required
  ## Metrics to collect from the XtremIO
  collectors = ["bbus","clusters","ssds","volumes","xms"]
```

### Metrics

- bbus
  - tags:
    - serial_number
    - guid
    - power_feed
    - name
    - model_name
  - fields:
    - bbus_power
    - bbus_average_daily_temp
    - bbus_enabled
    - bbus_ups_need_battery_replacement
    - bbus_ups_low_battery_no_input

- clusters
  - tags:
    - hardware_platform
    - license_id
    - guid
    - name
    - sys_psnt_serial_number
  - fields:
    - clusters_compression_factor
    - clusters_percent_memory_in_use
    - clusters_read_iops
    - clusters_write_iops
    - clusters_number_of_volumes
    - clusters_free_ssd_space_in_percent
    - clusters_ssd_num
    - clusters_data_reduction_ratio

- ssds
  - tags:
    - model_name
    - firmware_version
    - ssd_uid
    - guid
    - sys_name
    - serial_number
  - fields:
    - ssds_ssd_size
    - ssds_ssd_space_in_use
    - ssds_write_iops
    - ssds_read_iops
    - ssds_write_bandwidth
    - ssds_read_bandwidth
    - ssds_num_bad_sectors

- volumes
  - tags:
    - guid
    - sys_name
    - name
  - fields:
    - volumes_read_iops
    - volumes_write_iops
    - volumes_read_latency
    - volumes_write_latency
    - volumes_data_reduction_ratio
    - volumes_provisioned_space
    - volumes_used_space

- xms
  - tags:
    - guid
    - name
    - version
    - xms_ip
  - fields:
    - xms_write_iops
    - xms_read_iops
    - xms_overall_efficiency_ratio
    - xms_ssd_space_in_use
    - xms_ram_in_use
    - xms_ram_total
    - xms_cpu_usage_total
    - xms_write_latency
    - xms_read_latency
    - xms_user_accounts_count