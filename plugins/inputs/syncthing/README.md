# GitHub Input Plugin

Gather metrics from an instance of [Syncthing](https://syncthing.net) and the folders configured on the instance.

### Configuration

```toml
[[inputs.syncthing]]
  address = "http://localhost:8384"
  token = "1234asdf"
  http_timeout = "5s"
```

### Metrics

The exported metrics will look like this using the Prometheus output:

```
syncthing_folder_errors{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_global_bytes{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 4.2835204964e+10
syncthing_folder_global_deleted{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 49432
syncthing_folder_global_directories{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 30373
syncthing_folder_global_files{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 36981
syncthing_folder_global_symlinks{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_global_total_items{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 116786
syncthing_folder_ignore_patterns{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_in_sync_bytes{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 4.2835204964e+10
syncthing_folder_in_sync_files{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 36981
syncthing_folder_local_bytes{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 4.2835204964e+10
syncthing_folder_local_deleted{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 130
syncthing_folder_local_directories{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 30373
syncthing_folder_local_files{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 36981
syncthing_folder_local_symlinks{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_local_total_items{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 67484
syncthing_folder_need_bytes{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_need_deletes{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_need_directories{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_need_files{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_need_symlinks{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_need_total_items{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_pull_errors{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 0
syncthing_folder_sequence{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 484123
syncthing_folder_version{folder="xTDuT-kZeuK",host="kramacbook.local",instance="FN7I426-..."} 484123
syncthing_system_alloc{host="kramacbook.local",instance="FN7I426-..."} 1.6553996e+08
syncthing_system_cpu_percent{host="kramacbook.local",instance="FN7I426-..."} 1.6437177446884992
syncthing_system_folder_max_files{host="kramacbook.local",instance="FN7I426-..."} 185995
syncthing_system_folder_max_mib{host="kramacbook.local",instance="FN7I426-..."} 725111
syncthing_system_goroutines{host="kramacbook.local",instance="FN7I426-..."} 116
syncthing_system_memory_size{host="kramacbook.local",instance="FN7I426-..."} 16384
syncthing_system_memory_usage_mib{host="kramacbook.local",instance="FN7I426-..."} 360
syncthing_system_num_cpu{host="kramacbook.local",instance="FN7I426-..."} 8
syncthing_system_num_devices{host="kramacbook.local",instance="FN7I426-..."} 6
syncthing_system_num_folders{host="kramacbook.local",instance="FN7I426-..."} 4
syncthing_system_total_files{host="kramacbook.local",instance="FN7I426-..."} 227277
syncthing_system_total_mib{host="kramacbook.local",instance="FN7I426-..."} 853776
syncthing_system_uptime_seconds{host="kramacbook.local",instance="FN7I426-..."} 24158
```

### Example Output

```
> syncthing_folder,folder=9bjac-...,host=telegraf.local,instance=4XJDQDQ-... errors=0i,global_bytes=19663035411i,global_deleted=2543i,global_directories=24i,global_files=212i,global_symlinks=0i,global_total_items=2779i,ignore_patterns=false,in_sync_bytes=19663035411i,in_sync_files=212i,local_bytes=19663035411i,local_deleted=481i,local_directories=24i,local_files=212i,local_symlinks=0i,local_total_items=717i,need_bytes=0i,need_deletes=0i,need_directories=0i,need_files=0i,need_symlinks=0i,need_total_items=0i,pull_errors=0i,sequence=10507i,version=10507i 1584115043000000000
> syncthing_system,host=telegraf.local,instance=4XJDQDQ-... alloc=241845784i,cpu_percent=0.024645627195400695,folder_max_files=185995i,folder_max_mib=725111i,goroutines=107i,memory_size=3945i,memory_usage_mib=522i,num_cpu=2i,num_devices=6i,num_folders=4i,total_files=227277i,total_mib=853776i,uptime_seconds=3268505i 1584115044000000000
```

[syncthing]: https://www.syncthing.net
