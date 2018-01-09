# Linux Wireless Input Plugin

The Linux Wireless Plugin polls /proc/net/wireless for info and status on the Wireless network interfaces.
**This Plugin only works under Linux. A built-in OS-check exits on all other platforms.**
topic and adds messages to InfluxDB. This plugin allows a message to be in any of the supported `data_format` types.

## Configuration

```toml
# Read metrics from Wireless interface(s)
# dump_zeros will drop values that are zero
[[inputs.wireless]]
    proc_net_wireless = "/proc/net/wireless"
    dump_zeros = false
```

## Testing
The `wireless_test` mocks out the interaction with `/proc/net/wireless`. It requires no outside dependencies.
