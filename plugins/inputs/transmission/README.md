# Transmission Input Plugin

This plugin uses the [Transmission RPC API](https://github.com/transmission/transmission/blob/master/extras/rpc-spec.txt) (>= RPC Version 16 / Transmission 3.00) to gather information about the BitTorrent client and the status of it’s added torrents.

For this plugin to work the RPC API has to be enabled, and the `rpc-whitelist` has to be set accordingly if a remote connection is used. Further information can be found [here](https://github.com/transmission/transmission/wiki/Editing-Configuration-Files#rpc).

### Configuration:

```toml
# Collect Transmission client statistics about bandwidth usage and torrent status
[[inputs.transmission]]
  ## An URL where the Transmission RPC API is available
  url = "http://127.0.0.1:9091/transmission/rpc"
  
  ## Timeout for HTTP requests
  # timeout = "5s"
  
  ## Optional HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"
  
  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Metrics

- transmission
  - tags
    - url
    - rpc_host
    - rpc_port
    - peer_port
  - fields
    - torrents_active (integer)
    - torrents_stopped (integer)
    - torrents_queued_checking (integer)
    - torrents_checking (integer)
    - torrents_queued_downloading (integer)
    - torrents_downloading (integer)
    - torrents_queued_seeding (integer)
    - torrents_seeding (integer)
    - torrents_size (integer, total file size in bytes)
    - peers_connected (integer)
    - peers_getting_from_us (integer)
    - peers_sending_to_us (integer)
    - download_speed (integer, in B/s)
    - upload_speed (integer, in B/s)

### Example output:

```
transmission,peer_port=6881,rpc_host=127.0.0.1,rpc_port=9091,url=http://127.0.0.1:9091/transmission/rpc torrents_queued_checking=0i,torrents_checking=0i,torrents_size=834402528251i,peers_connected=63i,torrents_queued_seeding=0i,torrents_seeding=582i,peers_getting_from_us=16i,download_speed=0i,peers_sending_to_us=0i,torrents_active=1i,upload_speed=1577000i,torrents_stopped=30i,torrents_queued_downloading=0i,torrents_downloading=0i 1633606320000000000
```