# Linux Sysctl FS Input Plugin

The linux_sysctl_fs input provides Linux system level file metrics. The
documentation on these fields can be found at
<https://www.kernel.org/doc/Documentation/sysctl/fs.txt>.

Example output:

```shell
> linux_sysctl_fs,host=foo dentry-want-pages=0i,file-max=44222i,aio-max-nr=65536i,inode-preshrink-nr=0i,dentry-nr=64340i,dentry-unused-nr=55274i,file-nr=1568i,aio-nr=0i,inode-nr=35952i,inode-free-nr=12957i,dentry-age-limit=45i 1490982022000000000
```

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Provides Linux sysctl fs metrics
[[inputs.linux_sysctl_fs]]
  # no configuration
```
