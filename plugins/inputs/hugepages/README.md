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
  # types = ["root", "per_node"]
```

## Measurements

**The following measurements are supported by Hugepages plugin:**

- hugepages_root (gathered from root huge page control directory: `/sys/kernel/mm/hugepages`)
  - tags:
    - size_kb (integer, kB)
  - fields:
    - free (integer)
    - mempolicy (integer)
    - overcommit (integer)
    - reserved (integer)
    - surplus (integer)
    - total (integer)
- hugepages_per_node (gathered from per NUMA node directories: `/sys/devices/system/node/node[0-9]*/hugepages`)
  - tags:
    - size_kb (integer, kB)
    - node (integer)
  - fields:
    - free (integer)
    - surplus (integer)
    - total (integer)
- hugepages_meminfo (gathered from `/proc/meminfo` file)
  - The fields `total`, `free`, `reserved`, and `surplus` are counts of pages of default size. Fields with suffix `_kb` are in kilobytes.
  - fields:
    - anonymous_kb (integer, kB)
    - file_kb (integer, kB)
    - free (integer)
    - reserved (integer)
    - shared_kb (integer, kB)
    - size_kb (integer, kB)
    - surplus (integer)
    - tlb_kb (integer, kB)
    - total (integer)

## Example Output

```text
$ ./telegraf -config telegraf.conf -input-filter hugepages -test
> hugepages_root,host=ubuntu,size_kb=1048576 free=0i,mempolicy=8i,overcommit=0i,reserved=0i,surplus=0i,total=8i 1646258020000000000
> hugepages_root,host=ubuntu,size_kb=2048 free=883i,mempolicy=2048i,overcommit=0i,reserved=0i,surplus=0i,total=2048i 1646258020000000000
> hugepages_per_node,host=ubuntu,size_kb=1048576,node=0 free=0i,surplus=0i,total=4i 1646258020000000000
> hugepages_per_node,host=ubuntu,size_kb=2048,node=0 free=434i,surplus=0i,total=1024i 1646258020000000000
> hugepages_per_node,host=ubuntu,size_kb=1048576,node=1 free=0i,surplus=0i,total=4i 1646258020000000000
> hugepages_per_node,host=ubuntu,size_kb=2048,node=1 free=449i,surplus=0i,total=1024i 1646258020000000000
> hugepages_meminfo,host=ubuntu anonymous_kb=0i,file_kb=0i,free=883i,reserved=0i,shared_kb=0i,size_kb=2048i,surplus=0i,tlb_kb=12582912i,total=2048i 1646258020000000000
```
