# Telegraf plugin: NFS

The NFS plugin collects data from /proc/self/mountstats

### Configuration:

```toml
# NFS plugin for Telegraf
[[inputs.nfs]]
  # Collect only the common used metrics (bool)
  iostat = true
  # Collect all metrics available (bool)
  fullstat = true
```

### Measurements & Fields:

Measurement available only when the *fullstat* is set true
- nfs_events
    - attrinvalidates
    - congestionwait
    - datainvalidates
    - delay
    - dentryrevalidates
    - extendwrite
    - inoderevalidates
    - pnfsreads
    - pnfswrites
    - setattrtrunc
    - shortreads
    - shortwrites
    - sillyrenames
    - vfsflush
    - vfsfsync
    - vfslock
    - vfslookup
    - vfsopen
    - vfspermission
    - vfsreaddir
    - vfsreadpage
    - vfsreadpages
    - vfsrelease
    - vfssetattr
    - vfsupdatepage
    - vfswritepage
    - vfswritepages

- nfs_bytes
    - directreadbytes
    - directwritebytes
    - normalreadbytes
    - normalwritebytes
    - readpages
    - serverreadbytes
    - serverwritebytes
    - writepages

- nfs_xprttcp
    - backlogutil
    - badxids
    - bind_count
    - connect_count
    - connect_time
    - idle_time
    - inflightsends
    - rpcreceives
    - rpcsends

- nfs_ops
    - *_bytes_recv
    - *_bytes_sent
    - *_ops
    - *_queue_time
    - *_response_time
    - *_timeouts
    - *_total_time
    - *_trans

Measurement available only when the *iostat* is set true
- nfsstat_read
    - read_bytes
    - read_exe
    - read_ops
    - read_retrans
    - read_rtt

- nfsstat_write
    - write_bytes
    - write_exe
    - write_ops
    - write_retrans
    - write_rtt

### Tags:

- All measurements have the following tags:
    - mountpoint (Remote NFS addr + local mountpoint)

### Example Output:

```
$ telegraf -config /etc/telegraf/telegraf.conf -config-directory /etc/telegraf/telegraf.d -test -input-filter nfs
nfs_events,mountpoint=storage.local:/storage/\ /storage/library attrinvalidates=5096065,congestionwait=0,datainvalidates=1244054,delay=0,dentryrevalidates=2582462,extendwrite=1244042,inoderevalidates=3083569,pnfsreads=0,pnfswrites=0,setattrtrunc=0,shortreads=0,shortwrites=0,sillyrenames=315,vfsflush=2488084,vfsfsync=2488084,vfslock=0,vfslookup=1271789,vfsopen=3732142,vfspermission=15039496,vfsreaddir=26,vfsreadpage=0,vfsreadpages=1244044,vfsrelease=2488086,vfssetattr=0,vfsupdatepage=1244042,vfswritepage=1244042,vfswritepages=6220528 1469031319000000000
nfs_bytes,mountpoint=storage.local:/storage/\ /storage/library directreadbytes=0,directwritebytes=0,normalreadbytes=2608945209344,normalwritebytes=28612966,readpages=1244074,serverreadbytes=28744038,serverwritebytes=28612966,writepages=1244042 1469031319000000000
nfs_xprttcp,mountpoint=storage.local:/storage/\ /storage/library backlogutil=0,badxids=342,bind_count=1,connect_count=1,connect_time=0,idle_time=10,inflightsends=1203944825,rpcreceives=17934567,rpcsends=17934909 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library NULL_bytes_recv=0,NULL_bytes_sent=0,NULL_ops=0,NULL_queue_time=0,NULL_response_time=0,NULL_timeouts=0,NULL_total_time=0,NULL_trans=0 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library GETATTR_bytes_recv=345359504,GETATTR_bytes_sent=357674588,GETATTR_ops=3083563,GETATTR_queue_time=195462,GETATTR_response_time=3500483,GETATTR_timeouts=0,GETATTR_total_time=3767593,GETATTR_trans=3083568 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library SETATTR_bytes_recv=0,SETATTR_bytes_sent=0,SETATTR_ops=0,SETATTR_queue_time=0,SETATTR_response_time=0,SETATTR_timeouts=0,SETATTR_total_time=0,SETATTR_trans=0 1469031319000000000nfs_ops,mountpoint=storage.local:/storage/\ /storage/library LOOKUP_bytes_recv=159240484,LOOKUP_bytes_sent=174209048,LOOKUP_ops=1320593,LOOKUP_queue_time=4785,LOOKUP_response_time=919305,LOOKUP_timeouts=0,LOOKUP_total_time=950046,LOOKUP_trans=1320593 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library ACCESS_bytes_recv=533726520,ACCESS_bytes_sent=533524076,ACCESS_ops=4447720,ACCESS_queue_time=18490,ACCESS_response_time=4443524,ACCESS_timeouts=0,ACCESS_total_time=4557875,ACCESS_trans=4447720 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library READLINK_bytes_recv=0,READLINK_bytes_sent=0,READLINK_ops=0,READLINK_queue_time=0,READLINK_response_time=0,READLINK_timeouts=0,READLINK_total_time=0,READLINK_trans=0 1469031319000000000
nfsstat_read,mountpoint=storage.local:/storage/\ /storage/library read_bytes=348463344,read_exe=565842,read_ops=1244044,read_retrans=0,read_rtt=538537 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library READ_bytes_recv=189225712,READ_bytes_sent=159237632,READ_ops=1244044,READ_queue_time=11038,READ_response_time=538537,READ_timeouts=0,READ_total_time=565842,READ_trans=1244044 1469031319000000000
nfsstat_write,mountpoint=storage.local:/storage/\ /storage/library write_bytes=398093440,write_exe=2453883,write_ops=1244042,write_retrans=0,write_rtt=2419372 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library WRITE_bytes_recv=199046720,WRITE_bytes_sent=199046720,WRITE_ops=1244042,WRITE_queue_time=13147,WRITE_response_time=2419372,WRITE_timeouts=0,WRITE_total_time=2453883,WRITE_trans=1244042 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library CREATE_bytes_recv=338379424,CREATE_bytes_sent=204018396,CREATE_ops=1244042,CREATE_queue_time=4338,CREATE_response_time=3897411,CREATE_timeouts=0,CREATE_total_time=3930192,CREATE_trans=1244042 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library MKDIR_bytes_recv=0,MKDIR_bytes_sent=0,MKDIR_ops=0,MKDIR_queue_time=0,MKDIR_response_time=0,MKDIR_timeouts=0,MKDIR_total_time=0,MKDIR_trans=0 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library SYMLINK_bytes_recv=0,SYMLINK_bytes_sent=0,SYMLINK_ops=0,SYMLINK_queue_time=0,SYMLINK_response_time=0,SYMLINK_timeouts=0,SYMLINK_total_time=0,SYMLINK_trans=0 1469031319000000000nfs_ops,mountpoint=storage.local:/storage/\ /storage/library MKNOD_bytes_recv=0,MKNOD_bytes_sent=0,MKNOD_ops=0,MKNOD_queue_time=0,MKNOD_response_time=0,MKNOD_timeouts=0,MKNOD_total_time=0,MKNOD_trans=0 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library REMOVE_bytes_recv=179142048,REMOVE_bytes_sent=164214092,REMOVE_ops=1244042,REMOVE_queue_time=4737,REMOVE_response_time=2442833,REMOVE_timeouts=0,REMOVE_total_time=2474051,REMOVE_trans=1244042 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library RMDIR_bytes_recv=0,RMDIR_bytes_sent=0,RMDIR_ops=0,RMDIR_queue_time=0,RMDIR_response_time=0,RMDIR_timeouts=0,RMDIR_total_time=0,RMDIR_trans=0 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library RENAME_bytes_recv=81900,RENAME_bytes_sent=63000,RENAME_ops=315,RENAME_queue_time=2,RENAME_response_time=297,RENAME_timeouts=0,RENAME_total_time=306,RENAME_trans=315 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library LINK_bytes_recv=0,LINK_bytes_sent=0,LINK_ops=0,LINK_queue_time=0,LINK_response_time=0,LINK_timeouts=0,LINK_total_time=0,LINK_trans=0 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library READDIR_bytes_recv=0,READDIR_bytes_sent=0,READDIR_ops=0,READDIR_queue_time=0,READDIR_response_time=0,READDIR_timeouts=0,READDIR_total_time=0,READDIR_trans=0 1469031319000000000nfs_ops,mountpoint=storage.local:/storage/\ /storage/library READDIRPLUS_bytes_recv=15668,READDIRPLUS_bytes_sent=1820,READDIRPLUS_ops=13,READDIRPLUS_queue_time=0,READDIRPLUS_response_time=4,READDIRPLUS_timeouts=0,READDIRPLUS_total_time=4,READDIRPLUS_trans=13 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library FSSTAT_bytes_recv=480862368,FSSTAT_bytes_sent=323713936,FSSTAT_ops=2862494,FSSTAT_queue_time=125951,FSSTAT_response_time=9832796,FSSTAT_timeouts=0,FSSTAT_total_time=11515148,FSSTAT_trans=2862516 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library FSINFO_bytes_recv=328,FSINFO_bytes_sent=232,FSINFO_ops=2,FSINFO_queue_time=0,FSINFO_response_time=0,FSINFO_timeouts=0,FSINFO_total_time=0,FSINFO_trans=2 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library PATHCONF_bytes_recv=140,PATHCONF_bytes_sent=116,PATHCONF_ops=1,PATHCONF_queue_time=0,PATHCONF_response_time=0,PATHCONF_timeouts=0,PATHCONF_total_time=0,PATHCONF_trans=1 1469031319000000000
nfs_ops,mountpoint=storage.local:/storage/\ /storage/library COMMIT_bytes_recv=0,COMMIT_bytes_sent=0,COMMIT_ops=0,COMMIT_queue_time=0,COMMIT_response_time=0,COMMIT_timeouts=0,COMMIT_total_time=0,COMMIT_trans=0 1469031319000000000
```

#### References
[nfsiostat](http://git.linux-nfs.org/?p=steved/nfs-utils.git;a=summary)

[What is in /proc/self/mountstats for NFS mounts: an introduction](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex)
