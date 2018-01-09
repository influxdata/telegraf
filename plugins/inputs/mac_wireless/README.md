# Mac Wireless Input Plugin

The Mac Wireless Plugin polls /System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport -I for info and status on the Wireless network interfaces.
**This Plugin only works under Mac OS. A built-in OS-check exits on all other platforms.**
topic and adds messages to InfluxDB. This plugin allows a message to be in any of the supported `data_format` types.

## Configuration

```toml
# Read metrics from Wireless interface(s)
# dump_zeros will drop values that are zero
[[inputs.wireless]]
    cmd = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport -I"
    dump_zeros = false
```

## Testing
The `wireless_test` mocks out the interaction with `airport -I`. It requires no outside dependencies.
