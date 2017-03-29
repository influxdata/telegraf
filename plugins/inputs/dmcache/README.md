# DMCache Input Plugin

This plugin provide a native collection for dmsetup based statistics for dm-cache.

This plugin requires sudo, that is why you should setup and be sure that the telegraf is able to execute sudo without a password.

`sudo /sbin/dmsetup status --target cache` is the full command that telegraf will run for debugging purposes.

## Configuration

```
[[inputs.dmcache]]
  ## Whether to report per-device stats or not
  per_device = true
```
