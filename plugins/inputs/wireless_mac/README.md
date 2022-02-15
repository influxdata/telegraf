# Wireless Input Plugin for macOS

The wireless_mac plugin gathers metrics about wireless link quality file.
This plugin currently supports macOS only.

## Configuration

```toml
# Monitor wifi signal strength and quality
[[inputs.wireless]]
  ## Sets 'proc' directory path
  ## If not specified, then default is /proc
  # host_proc = "/proc"
```

## Metrics

- metric
  - tags:
    - host
    - interface (wireless interface, default: `airport`)
    - state (running or not )
    - op_mode (operating mode: station, ad_hoc or ap)
    - 802.11 Auth
    - link_auth
    - SSID (network identifier)
    - Channel
  - fields:
    - agrCtlRSSI
    - agrExtRSSI
    - agrCtlNoise
    - agrExtNoise
    - lastTxRate
    - maxRate
    - lastAssocStatus
    - MCS
    - NSS

## Example Output

This section shows example output in Line Protocol format.

```bash
wireless_mac,host=example.localdomain,interface=airport,state=running,op_mode=station,802.11_auth=open,link_auth=wpa2-psk,SSID=network_name,channel=153,80 MCS=8,NSS=3,agrCtlRSSI=-256i,agrCtlNoise=-256i,maxRate=450i,lastTxRate=100i,guardInterval=400 1519843022000000000
```
