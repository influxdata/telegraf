# MemInfo Input Plugin

The meminfo plugin collects system memory metrics from `/proc/meminfo`

### Configuration:
```toml
# Read metrics about memory usage from /proc/meminfo
[[inputs.meminfo]]
  # no configuration
```

### Metrics:
This collector gets _all_ stats from `/proc/meminfo`
It is possible that some stats won't be on a system because of the kernel in use

- meminfo
  - fields:
  	- Active (int)
  	- Active(anon) (int)
  	- Active(file) (int)
  	- AnonHugePages (int)
  	- AnonPages (int)
  	- Bounce (int)
  	- Buffers (int)
  	- Cached (int)
  	- CmaFree (int)
  	- CmaTotal (int)
  	- CommitLimit (int)
  	- Committed_AS (int)
  	- DirectMap1G (int)
  	- DirectMap2M (int)
  	- DirectMap4k (int)
  	- Dirty (int)
  	- HardwareCorrupted (int)
  	- HugePages_Free (int)
  	- HugePages_Rsvd (int)
  	- HugePages_Surp (int)
  	- HugePages_Total (int)
  	- Hugepagesize (int)
  	- Hugetlb (int)
  	- Inactive (int)
  	- Inactive(anon) (int)
  	- Inactive(file) (int)
  	- KernelStack (int)
  	- Mapped (int)
  	- MemAvailable (int)
  	- MemFree (int)
  	- MemTotal (int)
  	- Mlocked (int)
  	- NFS_Unstable (int)
  	- PageTables (int)
  	- Shmem (int)
  	- ShmemHugePages (int)
  	- ShmemPmdMapped (int)
  	- Slab (int)
  	- SReclaimable (int)
  	- SUnreclaim (int)
  	- SwapCached (int)
  	- SwapFree (int)
  	- SwapTotal (int)
  	- Unevictable (int)
  	- VmallocChunk (int)
  	- VmallocTotal (int)
  	- VmallocUsed (int)
  	- Writeback (int)
  	- WritebackTmp (int)

### Example Output:
```
meminfo SwapFree=8344563712i,Dirty=3026944i,SReclaimable=686764032i,HardwareCorrupted=0i,Hugepagesize=2097152i,Hugetlb=0i,Buffers=2527232i,SwapCached=0i,Active(file)=4796538880i,AnonPages=5512380416i,Bounce=0i,HugePages_Total=0i,DirectMap4k=476516352i,DirectMap1G=5368709120i,Active=7702962176i,Active(anon)=2906423296i,Unevictable=49152i,KernelStack=14745600i,CmaFree=0i,VmallocChunk=0i,Cached=9139920896i,Inactive(file)=5705285632i,Writeback=0i,PageTables=71294976i,NFS_Unstable=0i,Committed_AS=14902296576i,Inactive=6951710720i,CommitLimit=16623857664i,VmallocTotal=35184372087808i,AnonHugePages=0i,ShmemPmdMapped=0i,HugePages_Surp=0i,HugePages_Rsvd=0i,MemFree=788606976i,MemAvailable=11626881024i,Inactive(anon)=1246425088i,Mlocked=49152i,VmallocUsed=0i,CmaTotal=0i,MemTotal=16558592000i,Mapped=1402286080i,Shmem=1373892608i,Slab=900235264i,SUnreclaim=213471232i,WritebackTmp=0i,SwapTotal=8344563712i,ShmemHugePages=0i,HugePages_Free=0i,DirectMap2M=11072962560i 1533766374000000000

```
