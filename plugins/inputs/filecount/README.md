# filecount Input Plugin

Counts files in directories that match certain criteria.

### Configuration:

```toml
# Count files in a directory
[[inputs.filecount]]
  ## Directory to gather stats about.
  directory = "/var/cache/apt/archives"

  ## Only count files that match the name pattern. Defaults to "*".
  name = "*.deb"

  ## Count files in subdirectories. Defaults to true.
  recursive = false

  ## Only count regular files. Defaults to true.
  regular_only = true

  ## Only count files that are at least this size in bytes. If size is
  ## a negative number, only count files that are smaller than the
  ## absolute value of size. Defaults to 0.
  size = 0

  ## Only count files that have not been touched for at least this
  ## duration. If mtime is negative, only count files that have been
  ## touched in this duration. Defaults to "0s".
  mtime = "0s"
```

### Measurements & Fields:

- filecount
    - count (int)

### Tags:

- All measurements have the following tags:
    - directory (the directory path, as specified in the config)

### Example Output:

```
$ telegraf --config /etc/telegraf/telegraf.conf --input-filter filecount --test
> filecount,directory=/var/cache/apt,host=czernobog count=7i 1530034445000000000
> filecount,directory=/tmp,host=czernobog count=17i 1530034445000000000
```
