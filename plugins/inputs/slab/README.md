# Slab Input Plugin

This plugin collects details on how much memory each entry in Slab cache is
consuming. For example, it collects the consumption of `kmalloc-1024` and
`xfs_inode`. Since this information is obtained by parsing `/proc/slabinfo`
file, only Linux is supported. The specification of `/proc/slabinfo` has
not changed since [Linux v2.6.12 (April 2005)](https://github.com/torvalds/linux/blob/1da177e4/mm/slab.c#L2848-L2861),
so it can be regarded as sufficiently stable. The memory usage is
equivalent to the `CACHE_SIZE` column of `slabtop` command.

**Note: `/proc/slabinfo` is usually restricted to read as root user. Enable `use_sudo` option if necessary.**

## Configuration

```toml
# Get slab statistics from procfs
[[inputs.slab]]
  ## Use sudo to run LVM commands
  use_sudo = false
```

## Metrics

Metrics include generic ones such as `kmalloc_*` as well as those of kernel
subsystems and drivers used by the system such as `xfs_inode`.
Each field with `_size` suffix indicates memory consumption in bytes.

- mem
  - fields:
    _ kmalloc_8_size (integer)
    _ kmalloc_16_size (integer)
    _ kmalloc_32_size (integer)
    _ kmalloc_64_size (integer)
    _ kmalloc_96_size (integer)
    _ kmalloc_128_size (integer)
    _ kmalloc_256_size (integer)
    _ kmalloc_512_size (integer)
    _ xfs_ili_size (integer)
    _ xfs_inode_size (integer)

## Example Output

```shel
slab
kmalloc_1024_size=239927296i,kmalloc_512_size=5582848i 1651049129000000000
```
