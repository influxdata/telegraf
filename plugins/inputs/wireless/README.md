# Wireless Input Plugin

The wireless plugin gathers metrics about wireless link quality by reading the `/proc/net/wireless` file on linux and by executing the hidden `airport -I` command on macOS. This plugin currently supports linux and darwin (macOS) only.

## Configuration

```toml
# Monitor wifi signal strength and quality
[[inputs.wireless]]
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  ## Ignored on macOS/darwin
  # host_proc = "/proc"
```

The `host_proc` option is ignored on macOS.

## Metrics

### Linux metrics

- metric
  - tags:
    - interface (wireless interface)
  - fields:
    - status (int64, metric) - Its current state. This is a device dependent information
    - link (int64, percentage, metric) - general quality of the reception
    - level (int64, dBm, metric) - signal strength at the receiver
    - noise (int64, dBm, metric) - silence level (no packet) at the receiver
    - nwid (int64, packets, metric) - number of discarded packets due to invalid network id
    - crypt (int64, packets, metric) - number of packet unable to decrypt
    - frag (int64, packets, metric) - fragmented packets
    - retry (int64, packets, metric) - cumulative retry counts
    - misc (int64, packets, metric) - dropped for un-specified reason
    - missed_beacon (int64, packets, metric) - missed beacon packets


### macOS metrics

- metric
  - tags:
    - state (running or not)
    - op_mode (operating mode: station, ad_hoc or ap)
    - 802.11_auth (open or hidden)
    - link_auth (authorization scheme)

  - fields:
    - BSSID (string, mac address, _only_ reported if run as `root`)
    - SSID (string, The SSID of the network)
    - agrCtlRSSI (int64, dBm, metric) - The current aggregate RSSI of the link
    - agrExtRSSI (int64, dBm, metric) - The current aggregate external RSSI of the link
    - agrCtlNoise (int64, dBm, metric) - The current aggregate noise of the link
    - agrExtNoise (int64, dBm, metric) - The current aggregate external noise of the link
    - lastTxRate (int64, Mbps, metric) - The last transmit rate
    - maxRate (int64, Mbps, metric) - The maximum transmit rate
    - lastAssocStatus (int64, metric) - The last association status
    - MCS (int64, MCS, metric) - The last MCS
    - guardInterval (int64, guard, metric) - The guard interval
    - NSS (int64, NSS, metric) - The number of spatial streams
    - channel (channel information)
    - Interface (wireless interface, set to `airport`)
    - State (running or not)
    - Op mode (operating mode: station, ad_hoc or ap)
    - 802.11 auth (open or hidden)
    - link auth (authorization scheme)
    - BSSID (mac address, _only_ reported if run as `root`)
    - SSID (The SSID of the network)
    - channel (channel number)
  - fields:
    - BSSID (string, mac address, _only_ reported if run as `root`)
    - SSID (string, The SSID of the network)
    - agrCtlRSSI (int64, dBm, metric) - The current aggregate RSSI of the link
    - agrExtRSSI (int64, dBm, metric) - The current aggregate external RSSI of the link
    - agrCtlNoise (int64, dBm, metric) - The current aggregate noise of the link
    - agrExtNoise (int64, dBm, metric) - The current aggregate external noise of the link
    - lastTxRate (int64, Mbps, metric) - The last transmit rate
    - maxRate (int64, Mbps, metric) - The maximum transmit rate
    - lastAssocStatus (int64, metric) - The last association status
    - MCS (int64, MCS, metric) - The last MCS
    - guardInterval (int64, guard, metric) - The guard interval
    - NSS (int64, NSS, metric) - The number of spatial streams
    - channel (channel information)
  
## Example Output

This section shows example output in Line Protocol format.

### Linux output

```text
wireless,host=example.localdomain,interface=wlan0 misc=0i,frag=0i,link=60i,level=-50i,noise=-256i,nwid=0i,crypt=0i,retry=1525i,missed_beacon=0i,status=0i 1519843022000000000
```

### macOS output

**Note:** The `BSSID` field is only reported if the user is root.

```text
wireless,802.11_auth=open,host=mac.local,link_auth=wpa2-psk,op_mode=station,state=running BSSID="12:34:56:78:9a:bc",MCS=7i,NSS=3i,SSID="local_ssid",agrCtlNoise=-88i,agrCtlRSSI=-44i,agrExtNoise=0i,agrExtRSSI=0i,channel="153,80",guardInterval=800i,lastAssocStatus=0i,lastTxRate=975i,maxRate=217i 1645112457000000000
```
