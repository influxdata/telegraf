# Kernel Input Plugin

This plugin gathers metrics about the [Linux kernel][kernel] including, among
others, the [available entropy][entropy], [Kernel Samepage Merging][ksm] and
[Pressure Stall Information][psi].

⭐ Telegraf v0.11.0
🏷️ system
💻 linux

[kernel]: https://kernel.org/
[entropy]: https://www.kernel.org/doc/html/latest/admin-guide/sysctl/kernel.html#random
[ksm]: https://www.kernel.org/doc/html/latest/mm/ksm.html
[psi]: https://www.kernel.org/doc/html/latest/accounting/psi.html

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
  ## * psi - pressure stall information
  # collect = []
```

Please check the documentation of the underlying kernel interfaces in the
`/proc/stat` section of the [proc man page][man_proc], as well as in the
`/proc interfaces` section of the [random man page][man_random].

Kernel Samepage Merging is generally documented in the
[kernel documentation][ksm] and the available metrics exposed via sysfs
are documented in the [admin guide][ksm_admin].

Pressure Stall Information is exposed through `/proc/pressure` and is documented
in [kernel documentation][psi]. Kernel version 4.20+ is required.

[ksm_admin]: https://www.kernel.org/doc/html/latest/admin-guide/mm/ksm.html#ksm-daemon-sysfs-interface
[man_proc]: http://man7.org/linux/man-pages/man5/proc.5.html
[man_random]: https://man7.org/linux/man-pages/man4/random.4.html

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

- pressure (if `psi` is included in `collect`)
  - tags:
    - resource: cpu, memory, or io
    - type: some or full
  - floating-point fields: avg10, avg60, avg300
  - integer fields: total

## Example Output

Default config:

```text
kernel boot_time=1690487872i,context_switches=321398652i,entropy_avail=256i,interrupts=141868628i,processes_forked=946492i 1691339564000000000
```

If `ksm` is included in `collect`:

```text
kernel boot_time=1690487872i,context_switches=321252729i,entropy_avail=256i,interrupts=141783427i,ksm_full_scans=0i,ksm_max_page_sharing=256i,ksm_merge_across_nodes=1i,ksm_pages_shared=0i,ksm_pages_sharing=0i,ksm_pages_to_scan=100i,ksm_pages_unshared=0i,ksm_pages_volatile=0i,ksm_run=0i,ksm_sleep_millisecs=20i,ksm_stable_node_chains=0i,ksm_stable_node_chains_prune_millisecs=2000i,ksm_stable_node_dups=0i,ksm_use_zero_pages=0i,processes_forked=946467i 1691339522000000000
```

If `psi` is included in `collect`:

```text
pressure,resource=cpu,type=some avg10=1.53,avg60=1.87,avg300=1.73 1700000000000000000
pressure,resource=memory,type=some avg10=0.00,avg60=0.00,avg300=0.00 1700000000000000000
pressure,resource=memory,type=full avg10=0.00,avg60=0.00,avg300=0.00 1700000000000000000
pressure,resource=io,type=some avg10=0.0,avg60=0.0,avg300=0.0 1700000000000000000
pressure,resource=io,type=full avg10=0.0,avg60=0.0,avg300=0.0 1700000000000000000
pressure,resource=cpu,type=some total=1088168194i 1700000000000000000
pressure,resource=memory,type=some total=3463792i 1700000000000000000
pressure,resource=memory,type=full total=1429641i 1700000000000000000
pressure,resource=io,type=some total=68568296i 1700000000000000000
pressure,resource=io,type=full total=54982338i 1700000000000000000
```

Note that the combination for `resource=cpu,type=full` is omitted because it is
always zero.
