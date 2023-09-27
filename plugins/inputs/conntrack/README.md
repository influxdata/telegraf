# Conntrack Input Plugin

Collects stats from Netfilter's conntrack-tools.

There are two collection mechanisms for this plugin:

## /proc/net/stat/nf_conntrack

When a user specifies the `collect` config option with valid options, then the
plugin will loop through the files in `/proc/net/stat/nf_conntrack` to find
CPU specific values.

## Specific files and dirs

The second mechanism is for the user to specify a set of directories and files
to search through

At runtime, conntrack exposes many of those connection statistics within
`/proc/sys/net`. Depending on your kernel version, these files can be found in
either `/proc/sys/net/ipv4/netfilter` or `/proc/sys/net/netfilter` and will be
prefixed with either `ip` or `nf`.  This plugin reads the files specified
in its configuration and publishes each one as a field, with the prefix
normalized to ip_.

conntrack exposes many of those connection statistics within `/proc/sys/net`.
Depending on your kernel version, these files can be found in either
`/proc/sys/net/ipv4/netfilter` or `/proc/sys/net/netfilter` and will be
prefixed with either `ip_` or `nf_`.  This plugin reads the files specified
in its configuration and publishes each one as a field, with the prefix
normalized to `ip_`.

In order to simplify configuration in a heterogeneous environment, a superset
of directory and filenames can be specified.  Any locations that does nt exist
are ignored.

For more information on conntrack-tools, see the
[Netfilter Documentation](http://conntrack-tools.netfilter.org/).

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Collects conntrack stats from the configured directories and files.
# This plugin ONLY supports Linux
[[inputs.conntrack]]
  ## The following defaults would work with multiple versions of conntrack.
  ## Note the nf_ and ip_ filename prefixes are mutually exclusive across
  ## kernel versions, as are the directory locations.

  ## Look through /proc/net/stat/nf_conntrack for these metrics
  ## all - aggregated statistics
  ## percpu - include detailed statistics with cpu tag
  collect = ["all", "percpu"]

  ## User-specified directories and files to look through
  ## Directories to search within for the conntrack files above.
  ## Missing directories will be ignored.
  dirs = ["/proc/sys/net/ipv4/netfilter","/proc/sys/net/netfilter"]

  ## Superset of filenames to look for within the conntrack dirs.
  ## Missing files will be ignored.
  files = ["ip_conntrack_count","ip_conntrack_max",
          "nf_conntrack_count","nf_conntrack_max"]
```

## Metrics

A detailed explanation of each fields can be found in
[kernel documentation][kerneldoc]

[kerneldoc]: https://www.kernel.org/doc/Documentation/networking/nf_conntrack-sysctl.txt

- conntrack
  - `ip_conntrack_count` `(int, count)`: The number of entries in the conntrack table
  - `ip_conntrack_max` `(int, size)`: The max capacity of the conntrack table
  - `ip_conntrack_buckets`  `(int, size)`: The size of hash table.

With `collect = ["all"]`:

- `entries`: The number of entries in the conntrack table
- `searched`: The number of conntrack table lookups performed
- `found`: The number of searched entries which were successful
- `new`: The number of entries added which were not expected before
- `invalid`: The number of packets seen which can not be tracked
- `ignore`: The number of packets seen which are already connected to an entry
- `delete`: The number of entries which were removed
- `delete_list`: The number of entries which were put to dying list
- `insert`: The number of entries inserted into the list
- `insert_failed`: The number of insertion attempted but failed (same entry exists)
- `drop`: The number of packets dropped due to conntrack failure
- `early_drop`: The number of dropped entries to make room for new ones, if maxsize reached
- `icmp_error`: Subset of invalid. Packets that can't be tracked due to error
- `expect_new`: Entries added after an expectation was already present
- `expect_create`: Expectations added
- `expect_delete`: Expectations deleted
- `search_restart`: Conntrack table lookups restarted due to hashtable resizes

### Tags

With `collect = ["percpu"]` will include detailed statistics per CPU thread.

Without `"percpu"` the `cpu` tag will have `all` value.

## Example Output

```text
conntrack,host=myhost ip_conntrack_count=2,ip_conntrack_max=262144 1461620427667995735
```

with stats:

```text
conntrack,cpu=all,host=localhost delete=0i,delete_list=0i,drop=2i,early_drop=0i,entries=5568i,expect_create=0i,expect_delete=0i,expect_new=0i,found=7i,icmp_error=1962i,ignore=2586413402i,insert=0i,insert_failed=2i,invalid=46853i,new=0i,search_restart=453336i,searched=0i 1615233542000000000
conntrack,host=localhost ip_conntrack_count=464,ip_conntrack_max=262144 1615233542000000000
```
