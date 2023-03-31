# DMCache Input Plugin

This plugin provide a native collection for dmsetup based statistics for
dm-cache.

This plugin requires sudo, that is why you should setup and be sure that the
telegraf is able to execute sudo without a password.

`sudo /sbin/dmsetup status --target cache` is the full command that telegraf
will run for debugging purposes.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Provide a native collection for dmsetup based statistics for dm-cache
# This plugin ONLY supports Linux
[[inputs.dmcache]]
  ## Whether to report per-device stats or not
  per_device = true
```

## Metrics

- dmcache
  - length
  - target
  - metadata_blocksize
  - metadata_used
  - metadata_total
  - cache_blocksize
  - cache_used
  - cache_total
  - read_hits
  - read_misses
  - write_hits
  - write_misses
  - demotions
  - promotions
  - dirty

## Tags

- All measurements have the following tags:
  - device

## Example Output

```text
dmcache,device=example cache_blocksize=0i,read_hits=995134034411520i,read_misses=916807089127424i,write_hits=195107267543040i,metadata_used=12861440i,write_misses=563725346013184i,promotions=3265223720960i,dirty=0i,metadata_blocksize=0i,cache_used=1099511627776ii,cache_total=0i,length=0i,metadata_total=1073741824i,demotions=3265223720960i 1491482035000000000
```
