# Telegraf plugin: NFSClient

#### Plugin arguments:
- **fullstat** bool: Collect per-operation type metrics.  Defaults to false.
- **include_mounts** list(string): gather metrics for only these mounts.  Default is to watch all mounts.
- **exclude_mounts** list(string): gather metrics for all mounts, except those listed in this option. Excludes take precedence over includes.
- **include_operations** list(string): List of specific NFS operations to track.  See /proc/self/mountstats (the "per-op statistics" section) for complete lists of valid options for NFSv3 and NFSV4.  The default is to gather all metrics, but this is almost certainly *not* what you want (there are 22 operations for NFSv3, and well over 50 for NFSv4).  A suggested 'minimal' list of operations to collect for basic usage:  `['READ','WRITE','ACCESS','GETATTR','READDIR','LOOKUP','LOOKUP']`
- **exclude_operations** list(string): Gather all metrics, except those listed.  Excludes take precedence over includes.

#### Description

The NFSClient plugin collects data from /proc/self/mountstats. By default, only a limited number of general system-level metrics are collected, including `nfsstat_read` and `nfsstat_write`.
If `fullstat` is set, a great deal of additional metrics are collected, detailed below.

**NOTE** Many of the metrics, even if tagged with a mount point, are really _per-server_.  Thus, if you mount these two shares:  `nfs01:/vol/foo/bar` and `nfs01:/vol/foo/baz`, there will be two near identical entries in /proc/self/mountstats.


#### Examples

Example output for basic metrics showing server-wise read and write data:

```
nfsstat_read,mountpoint=/home,serverexport=nfs01:/home read_ops=9797i,read_retrans=0i,read_bytes=124i,read_rtt=7953i,read_exe=8200i 1608784749000000000
nfsstat_write,mountpoint=/home,serverexport=nfs01:/home write_exe=0i,write_ops=0i,w0rite_retrans=0i,write_bytes=0i,write_rtt=0i 1608784749000000000
```

Example output for `fullstat=true` metrics, which includes additional measurements for `nfs_bytes`, `nfs_events`, and `nfs_xprt_tcp` (and `nfs_xprt_udp` if present).
Additionally, per-OP metrics are collected, with examples for READ, LOOKUP, and NULL shown.
Please refer to `/proc/self/mountstats` for a list of supported NFS operations, as it changes as it changes periodically.

```
nfs_bytes,mountpoint=/home,serverexport=nfs01:/vol/home directreadbytes=0i,directwritebytes=0i,normalreadbytes=42648757667i,normalwritebytes=0i,readpages=10404603i,serverreadbytes=42617098139i,serverwritebytes=0i,writepages=0i 1608787697000000000
nfs_events,mountpoint=/home,serverexport=nfs01:/vol/home attrinvalidates=116i,congestionwait=0i,datainvalidates=65i,delay=0i,dentryrevalidates=5911243i,extendwrite=0i,inoderevalidates=200378i,pnfsreads=0i,pnfswrites=0i,setattrtrunc=0i,shortreads=0i,shortwrites=0i,sillyrenames=0i,vfsaccess=7203852i,vfsflush=117405i,vfsfsync=0i,vfsgetdents=3368i,vfslock=0i,vfslookup=740i,vfsopen=157281i,vfsreadpage=16i,vfsreadpages=86874i,vfsrelease=155526i,vfssetattr=0i,vfsupdatepage=0i,vfswritepage=0i,vfswritepages=215514i 1608787697000000000
nfs_xprt_tcp,mountpoint=/home,serverexport=nfs01:/vol/home backlogutil=0i,badxids=0i,bind_count=1i,connect_count=1i,connect_time=0i,idle_time=0i,inflightsends=15659826i,rpcreceives=2173896i,rpcsends=2173896i 1608787697000000000

nfs_ops,mountpoint=/home,serverexport=nfs01:/vol/home READ_bytes_recv=42783584732i,READ_bytes_sent=189594804i,READ_ops=1300676i,READ_queue_time=102795i,READ_response_time=1337335i,READ_timeouts=0i,READ_total_time=1447060i,READ_trans=1300676i,read_bytes=42973179536i,read_exe=1447060i,read_ops=1300676i,read_retrans=0i,read_rtt=1337335i 1608787697000000000
> nfs_ops,mountpoint=/home,serverexport=nfs01:/vol/home LOOKUP_bytes_recv=119608i,LOOKUP_bytes_sent=129824i,LOOKUP_ops=859i,LOOKUP_queue_time=5i,LOOKUP_response_time=130i,LOOKUP_timeouts=0i,LOOKUP_total_time=149i,LOOKUP_trans=859i 1608788310000000000
nfs_ops,mountpoint=/home,serverexport=nfs01:/vol/home NULL_bytes_recv=24i,NULL_bytes_sent=44i,NULL_ops=1i,NULL_queue_time=0i,NULL_response_time=0i,NULL_timeouts=0i,NULL_total_time=0i,NULL_trans=2i 1608787697000000000
```

#### References
1. [nfsiostat](http://git.linux-nfs.org/?p=steved/nfs-utils.git;a=summary)
2. [net/sunrpc/stats.c - Linux source code](https://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/tree/net/sunrpc/stats.c)
3. [What is in /proc/self/mountstats for NFS mounts: an introduction](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex)
4. [The xprt: data for NFS mounts in /proc/self/mountstats](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsXprt)



#### Measurements & Fields

Always collected:

- nfsstat_read
    - read_bytes (integer, bytes) - The number of bytes exchanged doing READ operations.  (sum bytes sent *and* received for WRITE Ops)
    - read_ops (integer, count) - The number of RPC READ operations executed.
    - read_retrans (integer, count) - The number of times an RPC READ operation had to be retried.
    - read_exe (integer, miliseconds) - The number of miliseconds it took to process the RPC READ operations.
    - read_rtt (integer, miliseconds) - The round-trip time for RPC READ operations.

- nfsstat_write
    - write_bytes (integer, bytes) - The number of bytes exchanged doing WRITE operations.  (sum bytes sent *and* received for WRITE Ops)
    - write_ops (integer, count) - The number of RPC WRITE operations executed.
    - write_retrans (integer, count) - The number of times an RPC WRITE operation had to be retried.
    - write_exe (integer, miliseconds) - The number of miliseconds it took to process the RPC WRITE operations.
    - write_rtt (integer, miliseconds) - The rount-trip time for RPC WRITE operations.

In addition enabling `fullstat` will make many more metrics available.

#### Tags

- All measurements have the following tags:
    - mountpoint The local mountpoint, for instance: "/var/www"
    - serverexport The full server export, for instance: "nfsserver.example.org:/export"



### Additional metrics

When `fullstat` is true, additional measurements are collected.  Tags are the same as above.

#### NFS Operations

Most descriptions come from Reference [[3](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex)] and `nfs_iostat.h`.  Field order is the same as in `/proc/self/mountstats` and in the Kernel source.

Please refer to `/proc/self/mountstats` for a list of supported NFS operations, as it changes as it changes periodically.

- nfs_bytes
    - fields:
        - normalreadbytes - (int, bytes) - Bytes read from the server via `read()`
        - normalwritebytes - (int, bytes) - Bytes written to the server via `write()`
        - directreadbytes - (int, bytes) - Bytes read with O_DIRECT set
        - directwritebytes - (int, bytes) -Bytes written with O_DIRECT set
        - serverreadbytes - (int, bytes) - Bytes read via NFS READ (via `mmap()`)
        - serverwritebytes - (int, bytes) - Bytes written via NFS WRITE (via `mmap()`)
        - readpages - (int, count) - Number of pages read
        - writepages - (int, count) - Number of pages written

- nfs_events - Per-event metrics
    - fields:
        - inoderevalidates - (int, count) - How many times cached inode attributes have to be re-validated from the server.
        - dentryrevalidates - (int, count) - How many times cached dentry nodes have to be re-validated.
        - datainvalidates - (int, count) - How many times an inode had its cached data thrown out.
        - attrinvalidates - (int, count) - How many times an inode has had cached inode attributes invalidated.
        - vfsopen - (int, count) - How many times files or directories have been `open()`'d.
        - vfslookup - (int, count) - How many name lookups in directories there have been.
        - vfsaccess - (int, count) - Number of calls to `access()`. (formerly called "vfspermission")

        - vfsupdatepage - (int, count) - Count of updates (and potential writes) to pages.
        - vfsreadpage - (int, count) - Number of pages read.
        - vfsreadpages - (int, count) - Count of how many times a _group_ of pages was read (possibly via `mmap()`?).
        - vfswritepage - (int, count) - Number of pages written.
        - vfswritepages - (int, count) - Count of how many times a _group_ of pages was written (possibly via `mmap()`?)
        - vfsgetdents - (int, count) - Count of directory entry reads with getdents(). These reads can be served from cache and don't necessarily imply actual NFS requests. (formerly called "vfsreaddir")
        - vfssetattr - (int, count) - How many times we've set attributes on inodes.
        - vfsflush - (int, count) - Count of times pending writes have been forcibly flushed to the server.
        - vfsfsync - (int, count) - Count of calls to `fsync()` on directories and files.
        - vfslock - (int, count) - Number of times a lock was attempted on a file (regardless of success or not).
        - vfsrelease - (int, count) - Number of calls to `close()`.
        - congestionwait - (int, count) - Believe unused by the Linux kernel, but it is part of the NFS spec.
        - setattrtrunc - (int, count) - How many times files have had their size truncated.
        - extendwrite - (int, count) - How many times a file has been grown because you're writing beyond the existing end of the file.
        - sillyrenames - (int, count) - Number of times an in-use file was removed (thus creating a temporary ".nfsXXXXXX" file)
        - shortreads - (int, count) - Number of times the NFS server returned less data than requested.
        - shortwrites - (int, count) - Number of times NFS server reports it wrote less data than requested.
        - delay - (int, count) - Occurances of EJUKEBOX ("Jukebox Delay", probably unused)
        - pnfsreads - (int, count) - Count of NFS v4.1+ pNFS reads.
        - pnfswrites - (int, count) - Count of NFS v4.1+ pNFS writes.

  - nfs_xprt_tcp
    - fields:
        - bind_count - (int, count) - Number of _completely new_ mounts to this server (sometimes 0?)
        - connect_count - (int, count) - How many times the client has connected to the server in question
        - connect_time - (int, jiffies) - How long the NFS client has spent waiting for its connection(s) to the server to be established.
        - idle_time - (int, seconds) - How long (in seconds) since the NFS mount saw any RPC traffic.
        - rpcsends - (int, count) - How many RPC requests this mount has sent to the server.
        - rpcreceives - (int, count) - How many RPC replies this mount has received from the server.
        - badxids - (int, count) - Count of XIDs sent by the server that the client doesn't know about.
        - inflightsends - (int, count) - Number of outstanding requests; always >1. (See reference #4 for comment on this field)
        - backlogutil - (int, count) - Cumulative backlog count

- nfs_xprt_udp
    - fields:
        - [same as nfs_xprt_tcp, except for connect_count, connect_time, and idle_time]

- nfs_ops
    - fields (In all cases, **"OP"** is replaced with the uppercase name of the NFS operation, _e.g._ "READ", "FSINFO", _etc_.  See /proc/self/mountstats for a full list.):
        - OP_ops - (int, count) - Total operations of this type.
        - OP_trans - (int, count) - Total transmissions of this type, including retransmissions: `OP_ops - OP_trans = total_retransmissions` (lower is better).
        - OP_timeouts - (int, count) - Number of major timeouts.
        - OP_bytes_sent - (int, count) - Bytes received, including headers (should also be close to on-wire size).
        - OP_bytes_recv - (int, count) - Bytes sent, including headers (should be close to on-wire size).
        - OP_queue_time - (int, milliseconds) - Cumulative time a request waited in the queue before sending this OP type.
        - OP_response_time - (int, milliseconds) - Cumulative time waiting for a response for this OP type.
        - OP_total_time - (int, milliseconds) - Cumulative time a request waited in the queue before sending.
        - OP_errors - (int, count) - Total number operations that complete with tk_status < 0 (usually errors).  This is a new field, present in kernel >=5.3, mountstats version 1.1

