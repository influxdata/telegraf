# Wireless Input Plugin

The wireless plugin gathers metrics about wireless link quality by reading the `/proc/net/wireless` file on linux and by executing the hidden `airport -I` command on macOS. This plugin currently supports linux and darwin (macOS) only.

## Configuration

```toml
# Monitor wifi signal strength and quality
[[inputs.wireless]]
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  # host_proc = "/proc"
  # dump_zeros = true | false
```

The `host_proc` option is ignored on macOS. The `dump_zeros` option is only read by macOS and will cause fields that have a `zero` value to be sent to influxdb.

## Metrics

### Linux metrics

- metric
  - tags:
    - interface (wireless interface)
  - fields:
    - status (int64, gauge) - Its current state. This is a device dependent information
    - link (int64, percentage, gauge) - general quality of the reception
    - level (int64, dBm, gauge) - signal strength at the receiver
    - noise (int64, dBm, gauge) - silence level (no packet) at the receiver
    - nwid (int64, packets, counter) - number of discarded packets due to invalid network id
    - crypt (int64, packets, counter) - number of packet unable to decrypt
    - frag (int64, packets, counter) - fragmented packets
    - retry (int64, packets, counter) - cumulative retry counts
    - misc (int64, packets, counter) - dropped for un-specified reason
    - missed_beacon (int64, packets, counter) - missed beacon packets


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
    - agrCtlRSSI (int64, dBm, gauge) - The current aggregate RSSI of the link
    - agrExtRSSI (int64, dBm, gauge) - The current aggregate external RSSI of the link
    - agrCtlNoise (int64, dBm, gauge) - The current aggregate noise of the link
    - agrExtNoise (int64, dBm, gauge) - The current aggregate external noise of the link
    - lastTxRate (int64, Mbps, gauge) - The last transmit rate
    - maxRate (int64, Mbps, gauge) - The maximum transmit rate
    - lastAssocStatus (int64, gauge) - The last association status
    - MCS (int64, MCS, gauge) - The last MCS

## Example Output

This section shows example output in Line Protocol format.

### linux output

```shell
wireless,host=example.localdomain,interface=wlan0 misc=0i,frag=0i,link=60i,level=-50i,noise=-256i,nwid=0i,crypt=0i,retry=1525i,missed_beacon=0i,status=0i 1519843022000000000
```

### macOS output

```bash
wireless_mac,host=example.localdomain,interface=airport,state=running,op_mode=station,802.11_auth=open,link_auth=wpa2-psk,SSID=network_name,channel=153,80 MCS=8,NSS=3,agrCtlRSSI=-256i,agrCtlNoise=-256i,maxRate=450i,lastTxRate=100i,guardInterval=400 1519843022000000000
```
