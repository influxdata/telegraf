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
- meminfo (gathered from `/proc/meminfo` file)
  - fields:
    - anonymous_kb (integer, kB)
    - default_size_kb (integer, kB)
    - file_kb (integer, kB)
    - free_of_default_size (integer)
    - reserved_of_default_size (integer)
    - shared_memory_kb (integer, kB)
    - surplus_of_default_size (integer)
    - total_consumed_by_all_sizes_kb (integer, kB)
    - total_of_default_size (integer)

## Example Output

```text
$ ./telegraf -config telegraf.conf -input-filter hugepages -test
> hugepages_root,host=ubuntu,size_kb=1048576 free=0i,mempolicy=8i,overcommit=0i,reserved=0i,surplus=0i,total=8i 1646258020000000000
> hugepages_root,host=ubuntu,size_kb=2048 free=883i,mempolicy=2048i,overcommit=0i,reserved=0i,surplus=0i,total=2048i 1646258020000000000
> hugepages_per_node,host=ubuntu,size_kb=1048576,node=0 free=0i,surplus=0i,total=4i 1646258020000000000
> hugepages_per_node,host=ubuntu,size_kb=2048,node=0 free=434i,surplus=0i,total=1024i 1646258020000000000
> hugepages_per_node,host=ubuntu,size_kb=1048576,node=1 free=0i,surplus=0i,total=4i 1646258020000000000
> hugepages_per_node,host=ubuntu,size_kb=2048,node=1 free=449i,surplus=0i,total=1024i 1646258020000000000
> hugepages_meminfo,host=ubuntu anonymous_kb=0i,default_size_kb=2048i,file_kb=0i,free_of_default_size=883i,reserved_of_default_size=0i,shared_memory_kb=0i,surplus_of_default_size=0i,total_consumed_by_all_sizes_kb=12582912i,total_of_default_size=2048i 1646258020000000000
```
