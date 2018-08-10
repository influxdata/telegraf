# Swap Input Plugin

The swap plugin collects system swap metrics.

For more information on what swap memory is, read [All about Linux swap space](https://www.linux.com/news/all-about-linux-swap-space).

### Configuration:

```toml
# Read metrics about swap memory usage
[[inputs.swap]]
  # no configuration
```

### Metrics:

- swap
  - fields:
    - free (int)
    - total (int)
    - used (int)
    - used_percent (float)
    - in (int)
    - out (int)

### Example Output:

```
swap total=20855394304i,used_percent=45.43883523785713,used=9476448256i,free=1715331072i 1511894782000000000
```
