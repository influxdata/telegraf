# Hugepages Input Plugin

The hugepages plugin gathers hugepages metrics including per NUMA node

## Configuration

```toml
# Collects hugepages metrics from kernel and per NUMA node
[[inputs.hugepages]]
  ## Path to a NUMA nodes
  # numa_node_path = "/sys/devices/system/node"
  ## Path to a meminfo file
  # meminfo_path = "/proc/meminfo"
```

## Measurements & Fields

- hugepages
  - free (int, kB)
  - nr (int, kB)
  - HugePages_Total (int, kB)
  - HugePages_Free (int, kB)

## Tags

- hugepages has the following tags:
  - node

## Example Output

```text
$ ./telegraf -config telegraf.conf -input-filter hugepages -test
> hugepages,host=maxpc,node=node0 free=0i,nr=0i 1467618621000000000
> hugepages,host=maxpc,name=meminfo HugePages_Free=0i,HugePages_Total=0i 1467618621000000000
```
