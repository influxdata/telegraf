# Telegraf plugin: NFSCLIENT

#### Plugin arguments:
- **fullstat** bool: Collect per-operation type metrics

#### Description

The NFSCLIENT plugin collects data from /proc/self/mountstats, by default it will only include a quite limited set of IO metrics. 
If fullstat is set, it will collect a lot of per-operation statistics.

#### Measurements & Fields

- nfsstat_read
    - read_bytes (integer, bytes) - The number of bytes exchanged doing READ operations.
    - read_ops (integer, count) - The number of RPC READ operations executed. 
    - read_retrans (integer, count) - The number of times an RPC READ operation had to be retried.
    - read_exe (integer, miliseconds) - The number of miliseconds it took to process the RPC READ operations.
    - read_rtt (integer, miliseconds) - The round-trip time for RPC READ operations.

- nfsstat_write
    - write_bytes (integer, bytes) - The number of bytes exchanged doing WRITE operations.
    - write_ops (integer, count) - The number of RPC WRITE operations executed.
    - write_retrans (integer, count) - The number of times an RPC WRITE operation had to be retried.
    - write_exe (integer, miliseconds) - The number of miliseconds it took to process the RPC WRITE operations.
    - write_rtt (integer, miliseconds) - The rount-trip time for RPC WRITE operations.

In addition enabling fullstat will make many more metrics available, but description of those is beyond the scope here.
See references for more details.

#### Tags

- All measurements have the following tags:
    - mountpoint The local mountpoint, for instance: "/var/www"
    - serverexport The full server export, for instance: "nfsserver.example.org:/export"

#### References
[nfsiostat](http://git.linux-nfs.org/?p=steved/nfs-utils.git;a=summary)
[net/sunrpc/stats.c - Linux source code](https://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/tree/net/sunrpc/stats.c)
[What is in /proc/self/mountstats for NFS mounts: an introduction](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex)
