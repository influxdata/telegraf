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

## Example Output
