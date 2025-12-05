# Bcache Input Plugin

This plugin gathers statistics for the [block layer cache][bcache]
from the `stats_total` directory and `dirty_data` file.

⭐ Telegraf v0.2.0
🏷️ system
💻 linux

[bcache]: https://docs.kernel.org/admin-guide/bcache.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics of bcache from stats_total and dirty_data
# This plugin ONLY supports Linux
[[inputs.bcache]]
  ## Bcache sets path
  ## If not specified, then default is:
  bcachePath = "/sys/fs/bcache"

  ## By default, Telegraf gather stats for all bcache devices
  ## Setting devices will restrict the stats to the specified
  ## bcache devices.
  bcacheDevs = ["bcache0"]
```

## Metrics

Tags:

- `backing_dev` device backed by the cache
- `bcache_dev` device used for caching

Fields:

- `dirty_data`: Amount of dirty data for this backing device in the cache.
  Continuously updated unlike the cache set's version, but may be slightly off
- `bypassed`: Amount of IO (both reads and writes) that has bypassed the cache
- `cache_bypass_hits`:  Hits for IO that is intended to skip the cache
- `cache_bypass_misses`:  Misses for IO that is intended to skip the cache
- `cache_hits`: Hits per individual IO as seen by bcache sees them; a
  partial hit is counted as a miss.
- `cache_misses`: Misses per individual IO as seen by bcache sees them; a
  partial hit is counted as a miss.
- `cache_hit_ratio`: Hit to miss ratio
- `cache_miss_collisions`: Instances where data was going to be inserted into
  cache from a miss, but raced with a write and data was already present
  (usually zero since the synchronization for cache misses was rewritten)
- `cache_readaheads`: Count of times readahead occurred.

## Example Output

```text
bcache,backing_dev="md10",bcache_dev="bcache0" dirty_data=11639194i,bypassed=5167704440832i,cache_bypass_hits=146270986i,cache_bypass_misses=0i,cache_hit_ratio=90i,cache_hits=511941651i,cache_miss_collisions=157678i,cache_misses=50647396i,cache_readaheads=0i
```
