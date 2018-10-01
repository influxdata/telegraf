# Wireless Input Plugin

The Wireless Plugin polls the wireless interface, if present, for info and status on the Wireless network interfaces.
**This Plugin only works under Linux or Mac OS X. A built-in OS-check exits on all other platforms.**

## Configuration

```toml
# Read metrics from Wireless interface(s)
# dump_zeros will drop values that are zero
[[inputs.wireless]]
    dump_zeros = false
```

## Testing
The `wireless_test` mocks out the interaction with the wireless interface. It requires no outside dependencies.