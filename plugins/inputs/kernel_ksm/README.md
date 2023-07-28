# Kernel KSM Input Plugin

This plugin is only available on Linux.

The kernel_ksm plugin gathers info about the kernel's KSM (Kernel Samepage
Merging) functionality.
Gathering metrics in this case works by picking up all useful values under
`/sys/kernel/mm/ksm`.

Kernel Samepage Merging is generally documented in [kernel documenation][1] and
the available metrics exposed via sysfs are documented in [admin guide][2]

[1]: https://www.kernel.org/doc/html/latest/mm/ksm.html
[2]: https://www.kernel.org/doc/html/latest/admin-guide/mm/ksm.html#ksm-daemon-sysfs-interface

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Get kernel statistics from /sys/kernel/mm/ksm
# This plugin ONLY supports Linux
[[inputs.kernel_ksm]]
  # no configuration
```

## Metrics

- kernel_ksm
  - full_scans (integer, how many times all mergeable areas have been scanned, `full_scans`)
  - max_page_sharing (integer, maximum sharing allowed for each KSM page, `max_page_sharing`)
  - merge_across_nodes (integer, whether pages should be merged across NUMA nodes, `merge_across_nodes`)
  - pages_shared (integer, how many shared pages are being used, `pages_shared`)
  - pages_sharing (integer,how many more sites are sharing them , `pages_sharing`)
  - pages_to_scan (integer, how many pages to scan before ksmd goes to sleep, `pages_to_scan`)
  - pages_unshared (integer, how many pages unique but repeatedly checked for merging, `pages_unshared`)
  - pages_volatile (integer, how many pages changing too fast to be placed in a tree, `pages_volatile`)
  - run (integer, whether ksm is running or not, `run`)
  - sleep_millisecs (integer, how many milliseconds ksmd should sleep between scans, `sleep_millisecs`)
  - stable_node_chains (integer, the number of KSM pages that hit the max_page_sharing limit, `stable_node_chains`)
  - stable_node_chains_prune_millisecs (integer, how frequently KSM checks the metadata of the pages that hit the deduplication limit, `stable_node_chains_prune_millisecs`)
  - stable_node_dups (integer, number of duplicated KSM pages, `stable_node_dups`)
  - use_zero_pages (integer, whether empty pages should be treated specially, `use_zero_pages`)

## Example Output

```text
kernel_ksm full_scans=58007i,max_page_sharing=256i,merge_across_nodes=1i,pages_shared=95572i,pages_sharing=181814i,pages_to_scan=1000i,pages_unshared=627454i,pages_volatile=15149i,run=1i,sleep_millisecs=20i,stable_node_chains=72i,stable_node_chains_prune_millisecs=2000i,stable_node_dups=576i,use_zero_pages=0i 1690634483000000000

```
