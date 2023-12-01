# QBittorrent Input Plugin

The qbittorrent plugin will gather and create metrics about the status of the torrent server.

Compatible qbittorrent API versions must be higher than 2.9.2. Lower versions might work but are neither tested nor recommended to use with this plugin.

## Global configuration options

## Configuration

```toml @sample.conf
# Read QBittorrent status information
[[inputs.qbittorrent]]
  ## Url of QBittorrent server
  # Tls can be:
  # url = "https://127.0.0.1:8080"
  url = "http://127.0.0.1:8080"

  ## Credentials for QBittorrent
  # username = "admin"
  # password = "admin"
```

## Metrics

- qbittorrent
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
  - tag_count
  - category_count

- torrent
  - added_on
  - amount_left
  - availability
  - completed
  - completion_on
  - download_limit
  - download_speed
  - downloaded
  - downloaded_session
  - eta
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
  - size
  - time_active
  - total_size
  - trackers_count
  - up_limit
  - uploaded
  - uploaded_session
  - upspeed

## Example Output

```text
qbittorrent,host=host all_time_download=200i,use_subcategories=false,queued_io_jobs=0i,up_info_speed=200i,queueing=true,connection_status="connected",dht_nodes=20i,free_space_on_disk=454373523i,dl_info_speed=100i,read_cache_hits="0",dl_rate_limit=0i,refresh_interval=1500i,global_ratio="0.32",total_queued_size=0i,all_time_upload=2134i,read_cache_overload="0",dl_info_data=343564i,up_info_data=29000i,write_cache_overload="0",use_alt_speed_limits=false,total_buffers_size=2020i,up_rate_limit=0i,total_wasted_session=755410338i,total_peer_connections=32i,average_time_queue=100i,category_count=31i,tag_count=10i,source="http://xxxx/xxx" 1700109890000000000
torrent,auto_tmm=false,content_path=/download/xxxxxxxxxxxxxxxx,fl_piece_prio=false,force_start=false,hash=xxxxxxxxxxx,host=SoberHoa-desktop,infohash_v1=xxxxxxxxxx,magnet_uri=magnet:?xt\=urn:btih:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,name=xxxxx,save_path=/download,seq_download=false,state=stalledUP,super_seeding=false,tags=tag1,tracker=https://xxxxxx.xx/announce.php?passkey\=xxxxxx uploaded=385341753i,max_seeding_time=-1i,amount_left=0i,num_leechs=1i,total_size=3694100172i,seen_complete=1700872684i,max_inactive_seeding_time=-1i,download_speed=0i,up_limit=0i,seeding_time=146659i,inactive_seeding_time_limit=-2i,downloaded_session=3698092707i,seeding_time_limit=-2i,priority=0i,progress=1,num_seeds=0i,ratio_limit=-2i,completed=3694100172i,trackers_count=1i,upspeed=0i,download_limit=0i,last_activity=1700844203i,availability=-1i,eta=8640000i,num_incomplete=204i,added_on=1700839030i,uploaded_session=645354i,time_active=43543,ratio=0.10420013329319708,max_ratio=-1i,num_complete=23i,completion_on=345i,downloaded=43534i,size=4546i,source="http://xxxx/xxx" 1700987990000000000
```
