# Syncthing Input Plugin

Gather metrics from an instance of [Syncthing](https://syncthing.net) and the folders configured on the instance.

### Configuration

```toml
[[inputs.syncthing]]
  address = "http://localhost:8384"
  token = "1234asdf"
  http_timeout = "5s"
```

### Metrics

- syncthing_system
  - tags:
    - host - The hostname
    - instance - The Syncthing instance UUID
  - fields:
    - alloc (integer)
    - cpu_percent (float, percentage)
    - folder_max_files (integer)
    - folder_max_mib (integer)
    - goroutines (integer)
    - memory_size (integer)
    - memory_usage_mib (integer)
    - num_cpu (integer)
    - num_devices (integer)
    - num_folders (integer)
    - total_files (integer)
    - total_mib (integer)
    - uptime_seconds (integer)
- syncthing_folder
  - tags:
    - host - The hostname
    - instance - The Syncthing instance UUID
    - folder - The folder ID
  - fields:
    - errors (integer)
    - global_bytes (integer)
    - global_deleted (integer)
    - global_directories (integer)
    - global_files (integer)
    - global_symlinks (integer)
    - global_total_items (integer)
    - ignore_patterns (bool)
    - in_sync_bytes (integer)
    - in_sync_files (integer)
    - local_bytes (integer)
    - local_deleted (integer)
    - local_directories (integer)
    - local_files (integer)
    - local_symlinks (integer)
    - local_total_items (integer)
    - need_bytes (integer)
    - need_deletes (integer)
    - need_directories (integer)
    - need_files (integer)
    - need_symlinks (integer)
    - need_total_items (integer)
    - pull_errors (integer)
    - sequence (integer)
    - version (integer)

### Troubleshooting

Syncthing will by default not listen to other hosts than `localhost`, so if you have trouble with connecting to your Syncthing, and you are not hosting telegraf on the same machine, you might need to make Syncthing accessable on the network.
Please be aware that this allows people on your network to access the Syncthing interface.

### Example Output

```
> syncthing_folder,folder=9bjac-...,host=telegraf.local,instance=4XJDQDQ-... errors=0i,global_bytes=19663035411i,global_deleted=2543i,global_directories=24i,global_files=212i,global_symlinks=0i,global_total_items=2779i,ignore_patterns=false,in_sync_bytes=19663035411i,in_sync_files=212i,local_bytes=19663035411i,local_deleted=481i,local_directories=24i,local_files=212i,local_symlinks=0i,local_total_items=717i,need_bytes=0i,need_deletes=0i,need_directories=0i,need_files=0i,need_symlinks=0i,need_total_items=0i,pull_errors=0i,sequence=10507i,version=10507i 1584115043000000000
> syncthing_system,host=telegraf.local,instance=4XJDQDQ-... alloc=241845784i,cpu_percent=0.024645627195400695,folder_max_files=185995i,folder_max_mib=725111i,goroutines=107i,memory_size=3945i,memory_usage_mib=522i,num_cpu=2i,num_devices=6i,num_folders=4i,total_files=227277i,total_mib=853776i,uptime_seconds=3268505i 1584115044000000000
```

[syncthing]: https://www.syncthing.net
