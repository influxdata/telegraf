# Hugepages Input Plugin

Transparent Huge Pages (THP) is a Linux memory management system that reduces the overhead of
Translation Lookaside Buffer (TLB) lookups on machines with large amounts of memory by using larger
memory pages.

Consult <https://www.kernel.org/doc/html/latest/admin-guide/mm/hugetlbpage.html> for more details.

## Configuration

```toml
# Gathers huge pages measurements.
[[inputs.hugepages]]
  ## Supported huge page types:
  ##   - "root" - based on root huge page control directory: /sys/kernel/mm/hugepages
  ##   - "per_node" - based on per NUMA node directories: /sys/devices/system/node/node[0-9]*/hugepages
  ##   - "meminfo" - based on /proc/meminfo file
  # hugepages_types = ["root", "per_node"]
```

## Measurements

**The following measurements are supported by Hugepages plugin:**

- hugepages_root (gathered from root huge page control directory: `/sys/kernel/mm/hugepages`)
  - tags:
    - hugepages_size_kb (integer, kB)
  - fields:
    - free_hugepages (integer)
    - nr_hugepages (integer)
    - nr_hugepages_mempolicy (integer)
    - nr_overcommit_hugepages (integer)
    - resv_hugepages (integer)
    - surplus_hugepages (integer)
- hugepages_per_node (gathered from per NUMA node directories: `/sys/devices/system/node/node[0-9]*/hugepages`)
  - tags:
    - hugepages_size_kb (integer, kB)
    - node (integer)
  - fields:
    - free_hugepages (integer)
    - nr_hugepages (integer)
    - surplus_hugepages (integer)
- meminfo (gathered from `/proc/meminfo` file)
  - fields:
    - AnonHugePages_kB (integer, kB)
    - ShmemHugePages_kB (integer, kB)
    - FileHugePages_kB (integer, kB)
    - HugePages_Total (integer)
    - HugePages_Rsvd (integer)
    - HugePages_Surp (integer)
    - HugePages_Free (integer)
    - Hugepagesize_kB (integer, kB)
    - Hugetlb_kB (integer, kB)

## Example Output

```text
$ ./telegraf -config telegraf.conf -input-filter hugepages -test
> hugepages_root,host=ubuntu,hugepages_size_kb=1048576 free_hugepages=0i,nr_hugepages=8i,nr_hugepages_mempolicy=8i,nr_overcommit_hugepages=0i,resv_hugepages=0i,surplus_hugepages=0i 1646258020000000000
> hugepages_root,host=ubuntu,hugepages_size_kb=2048 free_hugepages=883i,nr_hugepages=2048i,nr_hugepages_mempolicy=2048i,nr_overcommit_hugepages=0i,resv_hugepages=0i,surplus_hugepages=0i 1646258020000000000
> hugepages_per_node,host=ubuntu,hugepages_size_kb=1048576,node=0 free_hugepages=0i,nr_hugepages=4i,surplus_hugepages=0i 1646258020000000000
> hugepages_per_node,host=ubuntu,hugepages_size_kb=2048,node=0 free_hugepages=434i,nr_hugepages=1024i,surplus_hugepages=0i 1646258020000000000
> hugepages_per_node,host=ubuntu,hugepages_size_kb=1048576,node=1 free_hugepages=0i,nr_hugepages=4i,surplus_hugepages=0i 1646258020000000000
> hugepages_per_node,host=ubuntu,hugepages_size_kb=2048,node=1 free_hugepages=449i,nr_hugepages=1024i,surplus_hugepages=0i 1646258020000000000
> hugepages_meminfo,host=ubuntu AnonHugePages_kb=0i,FileHugePages_kb=0i,HugePages_Free=883i,HugePages_Rsvd=0i,HugePages_Surp=0i,HugePages_Total=2048i,Hugepagesize_kb=2048i,Hugetlb_kb=12582912i,ShmemHugePages_kb=0i 1646258020000000000

```
