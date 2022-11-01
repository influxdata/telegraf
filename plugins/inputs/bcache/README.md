# bcache Input Plugin

Get bcache stat from stats_total directory and dirty_data file.

## Metrics

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

## Description

```text
dirty_data
  Amount of dirty data for this backing device in the cache. Continuously
  updated unlike the cache set's version, but may be slightly off.

bypassed
  Amount of IO (both reads and writes) that has bypassed the cache


cache_bypass_hits
cache_bypass_misses
  Hits and misses for IO that is intended to skip the cache are still counted,
  but broken out here.

cache_hits
cache_misses
cache_hit_ratio
  Hits and misses are counted per individual IO as bcache sees them; a
  partial hit is counted as a miss.

cache_miss_collisions
  Counts instances where data was going to be inserted into the cache from a
  cache miss, but raced with a write and data was already present (usually 0
  since the synchronization for cache misses was rewritten)

cache_readaheads
  Count of times readahead occurred.
```

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Read metrics of bcache from stats_total and dirty_data
[[inputs.bcache]]
  ## Bcache sets path
  ## If not specified, then default is:
  bcachePath = "/sys/fs/bcache"

  ## By default, Telegraf gather stats for all bcache devices
  ## Setting devices will restrict the stats to the specified
  ## bcache devices.
  bcacheDevs = ["bcache0"]
```

## Example Output

```shell
./telegraf --config telegraf.conf --input-filter bcache --test
```

```shell
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
