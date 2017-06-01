# Entropy Plugin

Collects the current entropy pool size on a Linux OS.

### Configuration:

```toml
# # Reads the available entropy from /proc/sys/kernel/random/entropy_avail
# [[inputs.entropy]]
#   ## (Optional) Override the file from which to collect entropy stats. Default is:
#   proc = "/proc/sys/kernel/random/entropy_avail"
```

### Measurements & Fields:

- entropy
    - available (int): the number of bits of entropy currently available

### Tags:

This input does not use tags.

### Example Output:

```
$ telegraf -config telegraf.conf -test -input-filter entropy
* Plugin: entropy, Collection 1
> entropy,host=myhost available=880i 1462755667567437744
```