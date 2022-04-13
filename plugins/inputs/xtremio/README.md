# XtremIO Input Plugin

The `xtremio` plugin gathers metrics from a Dell EMC XtremIO Storage Array's V3 Rest API. Documentation can be found [here](https://dl.dell.com/content/docu96624_xtremio-storage-array-x1-and-x2-cluster-types-with-xms-6-3-0-to-6-3-3-and-xios-4-0-15-to-4-0-31-and-6-0-0-to-6-3-3-restful-api-3-x-guide.pdf?language=en_us)

## Configuration

```toml
 # Gathers Metrics From a Dell EMC XtremIO Storage Array's V3 API
[[inputs.xtremio]]
  ## XtremIO User Interface Endpoint
  url = "https://xtremio.example.com/" # required

  ## Credentials
  username = "user1"
  password = "pass123"

  ## Metrics to collect from the XtremIO
  # collectors = ["bbus","clusters","ssds","volumes","xms"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
```

## Metrics

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

## Example Output

> xio,guid=abcdefghifklmnopqrstuvwxyz111111,host=HOSTNAME,model_name=Eaton\ 5P\ 1550,name=X2-BBU,power_feed=PWR-B,serial_number=SER1234567890 bbus_average_daily_temp=22i,bbus_enabled=1i,bbus_power=286i,bbus_ups_low_battery_no_input=0i,bbus_ups_need_battery_replacement=0i 1638295340000000000
> xio,guid=abcdefghifklmnopqrstuvwxyz222222,host=HOSTNAME,model_name=Eaton\ 5P\ 1550,name=X1-BBU,power_feed=PWR-A,serial_number=SER1234567891 bbus_average_daily_temp=22i,bbus_enabled=1i,bbus_power=246i,bbus_ups_low_battery_no_input=0i,bbus_ups_need_battery_replacement=0i 1638295340000000000
> xio,guid=abcdefghifklmnopqrstuvwxyz333333,hardware_platform=X1,host=HOSTNAME,license_id=LIC123456789,name=SERVER01,sys_psnt_serial_number=FNM01234567890 clusters_compression_factor=1.5160012465000001,clusters_data_reduction_ratio=2.1613617899,clusters_free_ssd_space_in_percent=34i,clusters_number_of_volumes=36i,clusters_percent_memory_in_use=29i,clusters_read_iops=331i,clusters_ssd_num=50i,clusters_write_iops=4649i 1638295341000000000
