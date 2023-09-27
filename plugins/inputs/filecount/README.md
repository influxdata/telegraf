# Filecount Input Plugin

Reports the number and total size of files in specified directories.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Count files in a directory
[[inputs.filecount]]
  ## Directories to gather stats about.
  ## This accept standard unit glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/log/**    -> recursively find all directories in /var/log and count files in each directories
  ##   /var/log/*/*   -> find all directories with a parent dir in /var/log and count files in each directories
  ##   /var/log       -> count all files in /var/log and all of its subdirectories
  directories = ["/var/cache/apt", "/tmp"]

  ## Only count files that match the name pattern. Defaults to "*".
  name = "*"

  ## Count files in subdirectories. Defaults to true.
  recursive = true

  ## Only count regular files. Defaults to true.
  regular_only = true

  ## Follow all symlinks while walking the directory tree. Defaults to false.
  follow_symlinks = false

  ## Only count files that are at least this size. If size is
  ## a negative number, only count files that are smaller than the
  ## absolute value of size. Acceptable units are B, KiB, MiB, KB, ...
  ## Without quotes and units, interpreted as size in bytes.
  size = "0B"

  ## Only count files that have not been touched for at least this
  ## duration. If mtime is negative, only count files that have been
  ## touched in this duration. Defaults to "0s".
  mtime = "0s"
```

## Metrics

- filecount
  - tags:
    - directory (the directory path)
  - fields:
    - count (integer)
    - size_bytes (integer)
    - oldest_file_timestamp (int, unix time nanoseconds)
    - newest_file_timestamp (int, unix time nanoseconds)

## Example Output

```text
filecount,directory=/var/cache/apt count=7i,size_bytes=7438336i,oldest_file_timestamp=1507152973123456789i,newest_file_timestamp=1507152973123456789i 1530034445000000000
filecount,directory=/tmp count=17i,size_bytes=28934786i,oldest_file_timestamp=1507152973123456789i,newest_file_timestamp=1507152973123456789i 1530034445000000000
```
