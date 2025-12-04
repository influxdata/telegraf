# File statistics Input Plugin

This plugin gathers metrics about file existence, size, and other file
statistics.

⭐ Telegraf v0.13.0
🏷️ system
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read stats about given file(s)
[[inputs.filestat]]
  ## Files to gather stats about.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". See https://github.com/gobwas/glob.
  files = ["/etc/telegraf/telegraf.conf", "/var/log/**.log"]

  ## If true, read the entire file and calculate an md5 checksum.
  md5 = false
```

## Metrics

### Measurements & Fields

- filestat
  - exists (int, 0 | 1)
  - size_bytes (int, bytes)
  - modification_time (int, unix time nanoseconds)
  - md5 (optional, string)

### Tags

- All measurements have the following tags:
  - file (the path the to file, as specified in the config)

## Example Output

```text
filestat,file=/tmp/foo/bar,host=tyrion exists=0i 1507218518192154351
filestat,file=/Users/sparrc/ws/telegraf.conf,host=tyrion exists=1i,size=47894i,modification_time=1507152973123456789i  1507218518192154351
```
