# Slab Input Plugin

This plugin collects details on memory consumption of [Slab cache][slab] entries
by parsing the `/proc/slabinfo` file respecting the `HOST_PROC` environment
variable.

> [!NOTE]
> This plugin requires `/proc/slabinfo` to be readable by the Telegraf user.

⭐ Telegraf v1.23.0
🏷️ system
💻 linux

[slab]: https://www.kernel.org/doc/gorman/html/understand/understand011.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Get slab statistics from procfs
# This plugin ONLY supports Linux
[[inputs.slab]]
  # no configuration - please see the plugin's README for steps to configure
  # sudo properly
```

### Sudo configuration

Since the slabinfo file is only readable by root, the plugin runs
`sudo /bin/cat` to read the file.

Sudo can be configured to allow telegraf to run just the command needed to read
the slabinfo file. For example, if telegraf is running as the user `telegraf`
and `HOST_PROC` is not used, add this to the sudoers file

```text
telegraf ALL = (root) NOPASSWD: /bin/cat /proc/slabinfo
```

## Metrics

Metrics include generic ones such as `kmalloc_*` as well as those of kernel
subsystems and drivers used by the system such as `xfs_inode`.
Each field with `_size` suffix indicates memory consumption in bytes.

- mem
  - tags:
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

```text
slab kmalloc_1024_size=239927296i,kmalloc_512_size=5582848i 1651049129000000000
```
