# Slab Input Plugin

This plugin collects details on how much memory each entry in Slab cache is
consuming. For example, it collects the consumption of `kmalloc-1024` and
`xfs_inode`. Since this information is obtained by parsing `/proc/slabinfo`
file, only Linux is supported. The specification of `/proc/slabinfo` has
not changed since [Linux v2.6.12 (April 2005)](https://github.com/torvalds/linux/blob/1da177e4/mm/slab.c#L2848-L2861),
so it can be regarded as sufficiently stable. The memory usage is
equivalent to the `CACHE_SIZE` column of `slabtop` command.
If the HOST_PROC environment variable is set, Telegraf will use its value instead of `/proc`

**Note: `/proc/slabinfo` is usually restricted to read as root user. Make sure telegraf can execute `sudo` without password.**

## Configuration

```toml
# Get slab statistics from procfs
[[inputs.slab]]
  # no configuration - please see the plugin's README for steps to configure
  # sudo properly
```

## Sudo configuration

Since the slabinfo file is only readable by root, the plugin runs `sudo /bin/cat` to read the file.

Sudo can be configured to allow telegraf to run just the command needed to read the slabinfo file. For example, if telegraf is running as the user 'telegraf' and HOST_PROC is not used, add this to the sudoers file:
`telegraf ALL = (root) NOPASSWD: /bin/cat /proc/slabinfo`

## Metrics

Metrics include generic ones such as `kmalloc_*` as well as those of kernel
subsystems and drivers used by the system such as `xfs_inode`.
Each field with `_size` suffix indicates memory consumption in bytes.

- mem
  - fields:
    - kmalloc_8_size (integer)
    - kmalloc_16_size (integer)
    - kmalloc_32_size (integer)
    - kmalloc_64_size (integer)
    - kmalloc_96_size (integer)
    - kmalloc_128_size (integer)
    - kmalloc_256_size (integer)
    - kmalloc_512_size (integer)
    - xfs_ili_size (integer)
    - xfs_inode_size (integer)

## Example Output

```shel
slab
kmalloc_1024_size=239927296i,kmalloc_512_size=5582848i 1651049129000000000
```
