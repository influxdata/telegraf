# filecount Input Plugin

Counts files in directories that match certain criteria.

### Configuration:

```toml
# Count files in a directory and compute their size
[[inputs.filecount]]
  ## Directory to gather stats about.
  directory = "/var/cache/apt/archives"

  ## Also compute total size of matched elements. Defaults to false.
  count_size = false

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

  ## Output stats for every subdirectory. Defaults to false.
  recursive_print = false

  ## Only output directories whose sub elements weighs more than this
  ## size in bytes. Defaults to 0.
  recursive_print_size = 0
```

### Measurements & Fields:

- filecount
    - count (int)
    - size (int, in Bytes)

### Tags:

- All measurements have the following tags:
    - directory (the directory path)

### Example Output:

```
$ telegraf --config /etc/telegraf/telegraf.conf --input-filter filecount --test
> filecount,directory=/var/cache/apt,host=czernobog count=7i,size=7438336i 1530034445000000000
> filecount,directory=/tmp,host=czernobog count=17i,size=28934786i 1530034445000000000
```
