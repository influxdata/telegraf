# Wireless Input Plugin

The wireless plugin gathers metrics about wireless link quality by reading the `/proc/net/wireless` file. This plugin currently supports linux only.

### Configuration:

```toml
# Monitor wifi signal strength and quality
[[inputs.wireless]]
  ## file paths for proc files. If empty default paths will be used:
  ##    /proc/net/wireless
  proc_net_wireless = "/proc/net/wireless"

  ## dump metrics with 0 values too
  dump_zeros       = false
```

### Metrics:

- metric
  - tags:
    - interface (wireless interface)
  - fields:
    - status (int64) - Its current state. This is a device dependent information
    - link (int64, percentage) - general quality of the reception
    - level (int64, dBm) - signal strength at the receiver
    - noise (int64, dBm) - silence level (no packet) at the receiver
    - nwid (int64, packets) - number of discarded packets due to invalid network id
    - crypt (int64, packets) - number of packet unable to decrypt
    - frag (int64, packets) - fragmented packets
    - retry (int64, packets) - cumulative retry counts
    - misc (int64, packets) - dropped for un-specified reason
    - missed_beacon (int64, packets) - missed beacon packets

### Example Output:

This section shows example output in Line Protocol format.

```
wireless,host=example.localdomain,interface=wlan0 misc=0i,frag=0i,link=60i,level=-50i,noise=-256i,nwid=0i,crypt=0i,retry=1525i,missed_beacon=0i,status=0i 1519843022000000000
```
