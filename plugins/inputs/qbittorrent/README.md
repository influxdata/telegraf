# QBittorrent Input Plugin

The qbittorrent plugin will get torrents status.

Compatible qbittorrent API versions must be higher than 2.9.2.
Versions lower than 2.9.2 and higher than 2.1.1 are also compatible
in principle but have not been tested.

## Global configuration options

## Configuration

```toml @sample.conf
# Read Qbittorrent status information
[[inputs.qbittorrent]]
  ## Url of qbittorrent server
  # Tls can be:
  # url = "https://127.0.0.1:8080"
  url = "http://127.0.0.1:8080"

  ## Credentials for qbittorrent
  # username = "admin"
  # password = "admin"
```

## Metrics

- server_state
  - all_time_download
  - all_time_upload
  - average_time_queue
  - connection_status
  - dht_nodes
  - dl_info_data
  - dl_info_speed
  - dl_rate_limit
  - free_space_on_disk
  - global_ratio
  - queued_io_jobs
  - queueing
  - read_cache_hits
  - read_cache_overload
  - refresh_interval
  - total_buffers_size
  - total_peer_connections
  - total_queued_size
  - total_wasted_session
  - up_info_data
  - up_info_speed
  - up_rate_limit
  - use_alt_speed_limits
  - use_subcategories
  - write_cache_overload

- torrents
  - added_on
  - amount_left
  - auto_tmm
  - availability
  - completed
  - completion_on
  - download_limit
  - download_speed
  - downloaded
  - downloaded_session
  - eta
  - fl_piece_prio
  - force_start
  - inactive_seeding_time_limit
  - last_activity
  - max_inactive_seeding_time
  - max_ratio
  - max_seeding_time
  - num_complete
  - num_incomplete
  - num_leechs
  - num_seeds
  - priority
  - progress
  - ratio
  - ratio_limit
  - seeding_time
  - seeding_time_limit
  - seen_complete
  - seq_download
  - size
  - super_seeding
  - time_active
  - total_size
  - trackers_count
  - up_limit
  - uploaded
  - uploaded_session
  - upspeed

- tags
  - count

- category
  - count

## Example Output

```text
server_state,host=host all_time_download=200i,use_subcategories=false,queued_io_jobs=0i,up_info_speed=200i,queueing=true,connection_status="connected",dht_nodes=20i,free_space_on_disk=454373523i,dl_info_speed=100i,read_cache_hits="0",dl_rate_limit=0i,refresh_interval=1500i,global_ratio="0.32",total_queued_size=0i,all_time_upload=2134i,read_cache_overload="0",dl_info_data=343564i,up_info_data=29000i,write_cache_overload="0",use_alt_speed_limits=false,total_buffers_size=2020i,up_rate_limit=0i,total_wasted_session=755410338i,total_peer_connections=32i,average_time_queue=100i 1700109890000000000
torrents,content_path=/download/file_name,hash=xxxxx,host=host,infohash_v1=xxxxxxxx,magnet_uri=magnet:?xt\=urn:btih:xxxxxxx&dn\=xxxxxxx&tr\=https%3A%2F%2Ft.xxxx.xx%2Fannounce.php%3Fpasskey%xxxx,name=file_name,save_path=/download,state=stalledUP,tags=TAG1\,\ TAG2,tracker=https://xxx.xx/announce.php?passkey\=xxxxx num_complete=31i,seq_download=false,size=34643i,max_inactive_seeding_time=-1i,inactive_seeding_time_limit=-2i,num_seeds=0i,seeding_time_limit=-2i,download_limit=0i,auto_tmm=false,downloaded_session=0i,downloaded=436534i,fl_piece_prio=false,seen_complete=1699773762i,amount_left=0i,up_limit=0i,total_size=32142365i,availability=-1i,max_ratio=-1i,upspeed=0i,time_active=683889i,super_seeding=false,ratio=0.4420612080004777,trackers_count=1i,force_start=false,added_on=1699424219i,eta=8640000i,progress=1,num_leechs=0i,ratio_limit=-2i,num_incomplete=1i,completion_on=5436i,download_speed=0i,uploaded=34534i,uploaded_session=5754i,priority=0i,last_activity=235423i,completed=436543i,max_seeding_time=-1i,seeding_time=465436i 1700109890000000000
tags,host=host count=31i 1700109890000000000
category,host=host count=10i 1700109890000000000
```