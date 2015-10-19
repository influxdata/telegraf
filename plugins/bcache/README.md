# Telegraf plugin: bcache

Get bcache stat from stats_total directory and dirty_data file.

# Measurements

Meta:

- tags: `backing_dev=dev bcache_dev=dev`

Measurement names:

- dirty_data
- bypassed
- cache_bypass_hits
- cache_bypass_misses
- cache_hit_ratio
- cache_hits
- cache_miss_collisions
- cache_misses
- cache_readaheads

# Example output

Using this configuration:

```
[bcache]
  # Bcache sets path
  # If not specified, then default is:
  # bcachePath = "/sys/fs/bcache"
  #
  # By default, telegraf gather stats for all bcache devices
  # Setting devices will restrict the stats to the specified
  # bcache devices.
  # bcacheDevs = ["bcache0", ...]
```

When run with:

```
./telegraf -config telegraf.conf -filter bcache -test
```

It produces:

```
* Plugin: bcache, Collection 1
> [backing_dev="md10" bcache_dev="bcache0"] bcache_dirty_data value=11639194
> [backing_dev="md10" bcache_dev="bcache0"] bcache_bypassed value=5167704440832
> [backing_dev="md10" bcache_dev="bcache0"] bcache_cache_bypass_hits value=146270986
> [backing_dev="md10" bcache_dev="bcache0"] bcache_cache_bypass_misses value=0
> [backing_dev="md10" bcache_dev="bcache0"] bcache_cache_hit_ratio value=90
> [backing_dev="md10" bcache_dev="bcache0"] bcache_cache_hits value=511941651
> [backing_dev="md10" bcache_dev="bcache0"] bcache_cache_miss_collisions value=157678
> [backing_dev="md10" bcache_dev="bcache0"] bcache_cache_misses value=50647396
> [backing_dev="md10" bcache_dev="bcache0"] bcache_cache_readaheads value=0
```
