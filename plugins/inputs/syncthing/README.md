# Syncthing Input Plugin

The Syncthing input plugin collects information from one Syncthing API endpoint.

## Configuration:

```toml
## Syncthing host
url = "http://localhost:8384"

# token_file = "/path/to/file"
## OR
# token = "<api-access-token>"

## Optional TLS Config
# tls_ca = "/etc/telegraf/ca.pem"
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = false

## Amount of time allowed to complete the HTTP request
# timeout = "5s"
```

## Measurements collected

### syncthing_folder

Tags:
 * `id` - the folder ID
 * `label` - the label (name) you have assigned to the folder
 * `path` - the sync path on the local filesystem

Fields: 
 * `paused` - if the folder is paused
 * `needed` - how many files are needed to be in sync

### syncthing_device

Tags:
 * `device_id` - the device id that is connected
 * `name` - the name of the connected device 

Fields:
 * `address` - the IP address of the connected device
 * `client_version` - the version of Syncthing that the connected device is using
 * `connected` - if the device is connected
 * `crypto` - which version of TLS the device is using for encryption
 * `in_bytes_total` - how many bytes have been received by this device
 * `out_bytes_total` - how many bytes have been sent to this device
 * `paused` - if the device connection has been paused

### Example output

```
syncthing_connection,device_id=AIDPJXN-35PEWDU-IJKFJLV-BFELH5K-2PSCUKR-CNUEHI5-CIIDIA3-ITKWTUU,host=bilbos-laptop,name=samwise-desktop address="192.168.2.243:22000",client_version="v1.6.1",connected=true,crypto="TLS1.3-TLS_AES_128_CCM_SHA256",in_bytes_total=14i,out_bytes_total=14i,paused=false 1603647083000000000
syncthing_connection,device_id=USCWXIF-37NATKW-AXIYI4I-CDWV7L2-CXUTPUT-3V2URIM-ZPMADVS-2HBMTUA,host=bilbos-laptop,name=NAS address="192.168.2.91:22000",client_version="v1.4.0",connected=true,crypto="TLS1.3-TLS_AES_128_CCM_SHA256",in_bytes_total=314i,out_bytes_total=303i,paused=false 1603647083000000000
syncthing_folder,host=bilbos-laptop,id=rq7yw-sk82y,label=dotfiles,path=/home/bilbo/sync/dotfiles need=0i,paused=false 1603647083000000000
```
