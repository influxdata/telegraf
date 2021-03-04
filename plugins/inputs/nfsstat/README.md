# Kernel NFSStats Input Plugin

The kernel nfsstats plugin gathers counter data from the nfs client
module by reading /proc/net/rpc/nfs.

The counters are the same displayed by nfsstat(8). Currently only NFS
versions 3 and 4 are supported.

This code makes assumptions that the format of the NFS 3 and 4 counters will not
change in the near future, as they have been stable for many years, and
therefore uses the prefix for each stats line as an anchor including the number
of fields. Eg:

```
# grep '^proc' /proc/net/rpc/nfs
proc3 22 0 85660 0 2 5 0 0 0 0 0 0 0 0 0 0 0 0 10 1464 4 1 0
proc4 60 0 2248737 335600 0 42683 0 9826671 2 9752803 1676 3275 0 0 0 67 0 67 4816499 43971828 16963426 1182 1217 560 5 2 210 2093 14786 1042 3661802 5368 0 0 0 0 0 0 0 2878 574 573 97944 0 574 0 0 0 0 1182 0 67 0 0 573 0 0 0 0 0 0
```

The NFS version 3 and 4 counters are, respectively, represented by the fields in
the lines starting with proc3 and proc4. The second field is the number of
remaining fields, therefore the choice was made to use the prefixes as anchors
to parse the lines. This implies that if the number of fields changes the input
will no longer be captured.

```
const prefix_nfs3 string = "proc3 22 "
const prefix_nfs4 string = "proc4 60 "
```

The metrics have been deliberately created so that they can be aggregated as a
whole, and also by protocol version, and by operation type. A sample of some of
the operations can be seen below:

```
nfsstat_operations{host="nfsclient.example.com",nfsvers="3",op="read"} 0
nfsstat_operations{host="nfsclient.example.com",nfsvers="3",op="readdir"} 0
nfsstat_operations{host="nfsclient.example.com",nfsvers="3",op="readdirplus"} 10
nfsstat_operations{host="nfsclient.example.com",nfsvers="3",op="readlink"} 0
nfsstat_operations{host="nfsclient.example.com",nfsvers="3",op="remove"} 0
nfsstat_operations{host="nfsclient.example.com",nfsvers="3",op="rename"} 0
nfsstat_operations{host="nfsclient.example.com",nfsvers="3",op="write"} 0
nfsstat_operations{host="nfsclient.example.com",nfsvers="4",op="read"} 2.248939e+06
nfsstat_operations{host="nfsclient.example.com",nfsvers="4",op="readdir"} 3.662032e+06
nfsstat_operations{host="nfsclient.example.com",nfsvers="4",op="readlink"} 1042
nfsstat_operations{host="nfsclient.example.com",nfsvers="4",op="remove"} 1217
nfsstat_operations{host="nfsclient.example.com",nfsvers="4",op="rename"} 560
nfsstat_operations{host="nfsclient.example.com",nfsvers="4",op="write"} 335600
```

### Configuration:

```toml
# Get kernel statistics from nfs client
[[inputs.nfsstat]]
  # no configuration
```

### Measurements & Fields:

All measurements are int64 counters put under nfsstat_operations, which
correspond to IOPs sent to the NFS server. This exporter is not designed to
track latency, but rather volume of IOPs per type of IOP, which allows for a
clearer view of operation patterns accross the client fleet.

The measurements are global per nfs client code on the host, and therefore do
not have stats per filesystem nor per nfs server. These stats come from
mountstats and mountinfo, which are not in scope for this exporter.

The maximum cardinality of this exporter is expected to be the sum of fields on
NFS 3 and NFS 4 counters, which is currently 22 + 60 = 82.

### Tags:

Two tags are added to each measurement:

- nfsvers int64 - 3 or 4.
- op string     - one of the defined nfs operations defined for nfsvers.

### Example Output:

```
$ telegraf --config ~/telegraf.conf --input-filter nfsstat --test

> nfsstat,host=nfsclient.example.com,nfsvers=3,op=null operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=getattr operations=85660i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=setattr operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=lookup operations=2i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=access operations=5i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=readlink operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=read operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=write operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=create operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=mkdir operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=symlink operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=mknod operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=remove operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=rmdir operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=rename operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=link operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=readdir operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=readdirplus operations=10i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=fsstat operations=1464i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=fsinfo operations=4i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=pathconf operations=1i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=3,op=commit operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=null operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=read operations=2248939i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=write operations=335600i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=commit operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=open operations=42684i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=open_confirm operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=open_noattr operations=9827031i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=open_downgrade operations=2i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=close operations=9753163i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=setattr operations=1676i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=fsinfo operations=3278i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=renew operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=setclientid operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=setclientid_confirm operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=lock operations=67i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=lockt operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=locku operations=67i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=access operations=4817082i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=getattr operations=43972483i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=lookup operations=16964071i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=lookup_root operations=1183i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=remove operations=1217i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=rename operations=560i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=link operations=5i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=symlink operations=2i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=create operations=210i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=pathconf operations=2095i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=statfs operations=14786i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=readlink operations=1042i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=readdir operations=3662034i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=server_caps operations=5373i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=delegreturn operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=getacl operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=setacl operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=fs_locations operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=release_lockowner operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=secinfo operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=fsid_present operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=exchange_id operations=2883i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=create_session operations=575i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=destroy_session operations=574i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=sequence operations=97956i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=get_lease_time operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=reclaim_complete operations=575i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=layoutget operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=getdeviceinfo operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=layoutcommit operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=layoutreturn operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=secinfo_no_name operations=1183i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=test_stateid operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=free_stateid operations=67i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=getdevicelist operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=bind_conn_to_session operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=destroy_clientid operations=574i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=seek operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=allocate operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=deallocate operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=layoutstats operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=clone operations=0i 1611077696000000000
> nfsstat,host=nfsclient.example.com,nfsvers=4,op=copy operations=0i 1611077696000000000
```
