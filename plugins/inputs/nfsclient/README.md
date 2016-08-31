# Telegraf plugin: NFSCLIENT

#### Plugin arguments:
- **fullstat** bool: Collect per-operation type metrics

#### Description

The NFSCLIENT plugin collects data from /proc/self/mountstats, by default it will only include a quite limited set of IO metrics. 
If fullstat is set, it will collect a lot of per-operation statistics.

#### Measurements & Fields

- nfsstat_read
    - read_bytes (integer, bytes)
    - read_exe (integer, bytes)
    - read_ops (integer, bytes)
    - read_retrans (integer, miliseconds)
    - read_rtt (integer, miliseconds)

- nfsstat_write
    - write_bytes (integer, bytes)
    - write_ops (integer, bytes)
    - write_retrans (integer, bytes)
    - write_exe (integer, miliseconds)
    - write_rtt (integer, miliseconds)

In addition enabling fullstat will make many more metrics available.

#### Tags

- All measurements have the following tags:
    - mountpoint The local mountpoint, for instance: "/var/www"
    - serverexport The full server export, for instance: "nfsserver.example.org:/export"

#### References
[nfsiostat](http://git.linux-nfs.org/?p=steved/nfs-utils.git;a=summary)
[net/sunrpc/stats.c - Linux source code](https://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/tree/net/sunrpc/stats.c)

[What is in /proc/self/mountstats for NFS mounts: an introduction](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex)
