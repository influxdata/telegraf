# Telegraf plugin: NFSClient

#### Plugin arguments:
- **fullstat** bool: Collect per-operation type metrics

#### Description

The NFSClient plugin collects data from /proc/self/mountstats, by default it will only include a quite limited set of IO metrics.
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
1. [nfsiostat](http://git.linux-nfs.org/?p=steved/nfs-utils.git;a=summary)
2. [net/sunrpc/stats.c - Linux source code](https://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/tree/net/sunrpc/stats.c)
3. [What is in /proc/self/mountstats for NFS mounts: an introduction](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex)


### Additional metrics

When `fullstat` is true, these additional metrics are collected:

#### NFS Operations

Most descriptions come from Reference [[3](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex)] and `nfs_iostat.h`.  Field order is the same as in `/proc/self/mountstats` and in the Kernel source.

- nfs_bytes
    - tags:
        - mountpoint - The local mountpoint, for instance: "/var/www"
        - serverexport - The full server export, for instance: "nfsserver.example.org:/export"
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
    - tags:
        - mountpoint - The local mountpoint, for instance: "/var/www"
        - serverexport - The full server export, for instance: "nfsserver.example.org:/export"
    - fields:
        - inoderevalidates - (int, count) - How many times cached inode attributes have to be re-validated from the server.
        - dentryrevalidates - (int, count) - How many times cached dentry nodes have to be re-validated.
        - datainvalidates - (int, count) - How many times an inode had its cached data thrown out.
        - attrinvalidates - (int, count) - How many times an inode has had cached inode attributes invalidated.
        - vfsopen - (int, count) - How many times files or directories have been `open()`'d.
        - vfslookup - (int, count) - How many name lookups in directories there have been.
        - vfspermission - (int, count) - Number of calls to `access()`.
        - vfsupdatepage - (int, count) - Count of updates (and potential writes) to pages.
        - vfsreadpage - (int, count) - Number of pages read.
        - vfsreadpages - (int, count) - Count of how many times a _group_ of pages was read (possibly via `mmap()`?).
        - vfswritepage - (int, count) - Number of pages written.
        - vfswritepages - (int, count) - Count of how many times a _group_ of pages was written (possibly via `mmap()`?)
        - vfsreaddir - (int, count) - Count of directory entry reads with getdents(). These reads can be served from cache and don't necessarily imply actual NFS requests.
        - vfssetattr - (int, count) - How many times we've set attributes on inodes.
        - vfsflush - (int, count) - Count of times pending writes have been forcibly flushed to the server.
        - vfsfsync - (int, count) - Count of calls to `fsync()` on directories and files.
        - vfslock - (int, count) - Number of times a lock was attempted on a file (regardless of success or not).
        - vfsrelease - (int, count) - Number of calls to `close()`.
        - congestionwait - (int, count) - Unused.
        - setattrtrunc - (int, count) - How many times files have had their size truncated.
        - extendwrite - (int, count) - How many times a file has been grown because you're writing beyond the existing end of the file.
        - sillyrenames - (int, count) - Number of times an in-use file was removed (thus creating a temporary ".nfsXXXXXX" file)
        - shortreads - (int, count) - Number of times the NFS server returned less data than requested.
        - shortwrites - (int, count) - Number of times NFS server reports it wrote less data than requested.
        - delay - (int, count) - Occurances of EJUKEBOX ("Jukebox Delay", probably unused)
        - pnfsreads - (int, count) - Count of NFS v4.1+ pNFS reads.
        - pnfswrites - (int, count) - Count of NFS v4.1+ pNFS writes.

  - nfs_xprt_tcp
    - tags:
        - mountpoint - The local mountpoint, for instance: "/var/www"
        - serverexport - The full server export, for instance: "nfsserver.example.org:/export"
    - fields:
        - bind_count - (int, count) - Number of _completely new_ mounts to this server (sometimes 0?)
        - connect_count - (int, count) - How many times the client has connected to the server in question
        - connect_time - (int, jiffies) - How long the NFS client has spent waiting for its connection(s) to the server to be established.
        - idle_time - (int, seconds) - How long (in seconds) since the NFS mount saw any RPC traffic.
        - rpcsends - (int, count) - How many RPC requests this mount has sent to the server.
        - rpcreceives - (int, count) - How many RPC replies this mount has received from the server.
        - badxids - (int, count) - Count of XIDs sent by the server that the client doesn't know about.
        - inflightsends - (int, count) - Number of outstanding requests; always >1.  ("Every time we send a request, we add the current difference between sends and receives to this number. Since we've just sent a request, this is always going to be at least one. This number is not as useful as you think it should be.")
        - backlogutil - (int, count) - Cumulative backlog count

- nfs_xprt_udp
    - tags:
        - [same as nfs_xprt_tcp]
    - fields:
        - [same as nfs_xprt_tcp, except for connect_count, connect_time, and idle_time]

- nfs_ops
    - tags:
        - mountpoint - The local mountpoint, for instance: "/var/www"
        - serverexport - The full server export, for instance: "nfsserver.example.org:/export"
    -fields - in all cases, "OP" is replaced with the uppercase name of the NFS operation, (_e.g._ "READ", "FSINFO", etc)
        - OP_ops - (int, count) - Total operations of this type.
        - OP_trans - (int, count) - Total transmissions of this type, including retransmissions: OP_ops - OP_trans = total_retransmissions (lower is better).
        - OP_timeouts - (int, count) - Number of major timeouts.
        - OP_bytes_sent - (int, count) - Bytes received, including headers (should also be close to on-wire size).
        - OP_bytes_recv - (int, count) - Bytes sent, including headers (should be close to on-wire size).
        - OP_queue_time - (int, milliseconds) - Cumulative time a request waited in the queue before sending this OP type.
        - OP_response_time - (int, milliseconds) - Cumulative time waiting for a response for this OP type.
        - OP_total_time - (int, milliseconds) - Cumulative time a request waited in the queue before sending.
        - OP_errors - (int, count) - Total number operations that complete with tk_status < 0 (usually errors).  This is a new field, present in kernel >=5.3, mountstats version 1.1
