# Kernel Input Plugin

This plugin is only available on Linux.

The kernel plugin gathers info about the kernel that doesn't fit into other
plugins. In general, it is the statistics available in `/proc/stat` that are not
covered by other plugins as well as the value of
`/proc/sys/kernel/random/entropy_avail` and optionally, Kernel Samepage Merging.

The metrics are documented in `man proc` under the `/proc/stat` section.
The metrics are documented in `man 4 random` under the `/proc/stat` section.

```text

/proc/sys/kernel/random/entropy_avail
Contains the value of available entropy

/proc/stat
kernel/system statistics. Varies with architecture. Common entries include:

page 5741 1808
The number of pages the system paged in and the number that were paged out (from disk).

swap 1 0
The number of swap pages that have been brought in and out.

intr 1462898
This line shows counts of interrupts serviced since boot time, for each of
the possible system interrupts. The first column is the total of all
interrupts serviced; each subsequent column is the total for a particular interrupt.

ctxt 115315
The number of context switches that the system underwent.

btime 769041601
boot time, in seconds since the Epoch, 1970-01-01 00:00:00 +0000 (UTC).

processes 86031
Number of forks since boot.
```

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
# Plugin to collect various Linux kernel statistics.
# This plugin ONLY supports Linux
[[inputs.kernel]]
  ## Additional gather options
  ## Possible options include:
  ## * ksm - kernel same-page merging
  # collect = []
```

## Metrics

- kernel
  - boot_time (integer, seconds since epoch, `btime`)
  - context_switches (integer, `ctxt`)
  - disk_pages_in (integer, `page (0)`)
  - disk_pages_out (integer, `page (1)`)
  - interrupts (integer, `intr`)
  - processes_forked (integer, `processes`)
  - entropy_avail (integer, `entropy_available`)
  - ksm_full_scans (integer, how many times all mergeable areas have been scanned, `full_scans`)
  - ksm_max_page_sharing (integer, maximum sharing allowed for each KSM page, `max_page_sharing`)
  - ksm_merge_across_nodes (integer, whether pages should be merged across NUMA nodes, `merge_across_nodes`)
  - ksm_pages_shared (integer, how many shared pages are being used, `pages_shared`)
  - ksm_pages_sharing (integer,how many more sites are sharing them , `pages_sharing`)
  - ksm_pages_to_scan (integer, how many pages to scan before ksmd goes to sleep, `pages_to_scan`)
  - ksm_pages_unshared (integer, how many pages unique but repeatedly checked for merging, `pages_unshared`)
  - ksm_pages_volatile (integer, how many pages changing too fast to be placed in a tree, `pages_volatile`)
  - ksm_run (integer, whether ksm is running or not, `run`)
  - ksm_sleep_millisecs (integer, how many milliseconds ksmd should sleep between scans, `sleep_millisecs`)
  - ksm_stable_node_chains (integer, the number of KSM pages that hit the max_page_sharing limit, `stable_node_chains`)
  - ksm_stable_node_chains_prune_millisecs (integer, how frequently KSM checks the metadata of the pages that hit the deduplication limit, `stable_node_chains_prune_millisecs`)
  - ksm_stable_node_dups (integer, number of duplicated KSM pages, `stable_node_dups`)
  - ksm_use_zero_pages (integer, whether empty pages should be treated specially, `use_zero_pages`)

## Example Output

```text
kernel boot_time=1690487872i,context_switches=321398652i,entropy_avail=256i,interrupts=141868628i,processes_forked=946492i 1691339564000000000

kernel boot_time=1690487872i,context_switches=321252729i,entropy_avail=256i,interrupts=141783427i,ksm_full_scans=0i,ksm_max_page_sharing=256i,ksm_merge_across_nodes=1i,ksm_pages_shared=0i,ksm_pages_sharing=0i,ksm_pages_to_scan=100i,ksm_pages_unshared=0i,ksm_pages_volatile=0i,ksm_run=0i,ksm_sleep_millisecs=20i,ksm_stable_node_chains=0i,ksm_stable_node_chains_prune_millisecs=2000i,ksm_stable_node_dups=0i,ksm_use_zero_pages=0i,processes_forked=946467i 1691339522000000000

```
