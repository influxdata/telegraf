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

| field        | type  | descripton                                                              |
|--------------|:-----:|-------------------------------------------------------------------------|
| free         | int   | free swap memory in bytes                                               |
| total        | int   | total swap memory in bytes                                              |
| used         | int   | used swap memory in bytes                                               |
| used_percent | float | percentage of swap memory used                                          |
| in           | int   | data swapped in since last boot in bytes (calculated from page number)  |
| out          | int   | data swapped out since last boot in bytes (calculated from page number) |

### Example Output:

```
swap total=20855394304i,used_percent=45.43883523785713,used=9476448256i,free=1715331072i 1511894782000000000
```
