# NSD Input Plugin

This plugin gathers stats from [NSD](https://nlnetlabs.nl/projects/nsd/about/) - an authoritative DNS server.

### Configuration

```toml
 # A plugin to collect stats from the NSD DNS server
 [[inputs.nsd]]
   ## Address of server to connect to, read from nsd conf default, optionally ':port'
   ## Will lookup IP if given a hostname
   server = "127.0.0.1:8952"
 
   ## If running as a restricted user you can prepend sudo for additional access:
   # use_sudo = false
 
   ## The default location of the nsd-control binary can be overridden with:
   # binary = "/usr/sbin/nsd-control"
 
   ## The default timeout of 1s can be overriden with:
   # timeout = "1s"
 
   ## When set to true, thread metrics are tagged with the thread id.
   ##
   ## The default is false for backwards compatibility, and will be change to
   ## true in a future version.  It is recommended to set to true on new
   ## deployments.
   thread_as_tag = false
```

#### Permissions:

It's important to note that this plugin references nsd-control, which may require additional permissions to execute successfully.
Depending on the user/group permissions of the telegraf user executing this plugin, you may need to alter the group membership, set facls, or use sudo.

**Group membership (Recommended)**:
```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G nsd telegraf

$ groups telegraf
telegraf : telegraf nsd
```

**Sudo privileges**:
If you use this method, you will need the following in your telegraf config:
```toml
[[inputs.nsd]]
  use_sudo = true
```

You will also need to update your sudoers file:
```bash
$ visudo
# Add the following line:
telegraf ALL=(ALL) NOPASSWD: /usr/sbin/nsd-control
```

Please use the solution you see as most appropriate.

### Metrics:

This is the full list of stats provided by nsd-control and potentially collected
depending of your nsd configuration. In the output, the dots in the nsd-control stat name are replaced by underscores(see
https://nlnetlabs.nl/documentation/nsd/nsd-control for details).

Shown metrics are with `thread_as_tag` enabled.

- nsd
  - fields:
    - num.queries
    - time.boot
    - time.elapsed
    - size.db.disk
    - size.db.mem
    - size.xfrd.mem
    - size.config.disk
    - size.config.mem
    - num.type.A
    - num.type.NS
    - num.type.MD
    - num.type.MF
    - num.type.CNAME
    - num.type.SOA
    - num.type.MB
    - num.type.MG
    - num.type.MR
    - num.type.NULL
    - num.type.WKS
    - num.type.PTR
    - num.type.HINFO
    - num.type.MINFO
    - num.type.MX
    - num.type.TXT
    - num.type.RP
    - num.type.AFSDB
    - num.type.X25
    - num.type.ISDN
    - num.type.RT
    - num.type.NSAP
    - num.type.SIG
    - num.type.KEY
    - num.type.PX
    - num.type.AAAA
    - num.type.LOC
    - num.type.NXT
    - num.type.SRV
    - num.type.NAPTR
    - num.type.KX
    - num.type.CERT
    - num.type.DNAME
    - num.type.OPT
    - num.type.APL
    - num.type.DS
    - num.type.SSHFP
    - num.type.IPSECKEY
    - num.type.RRSIG
    - num.type.NSEC
    - num.type.DNSKEY
    - num.type.DHCID
    - num.type.NSEC3
    - num.type.NSEC3PARAM
    - num.type.TLSA
    - num.type.SMIMEA
    - num.type.CDS
    - num.type.CDNSKEY
    - num.type.OPENPGPKEY
    - num.type.CSYNC
    - num.type.SPF
    - num.type.NID
    - num.type.L32
    - num.type.L64
    - num.type.LP
    - num.type.EUI48
    - num.type.EUI64
    - num.opcode.QUERY
    - num.class.IN
    - num.rcode.NOERROR
    - num.rcode.FORMERR
    - num.rcode.SERVFAIL
    - num.rcode.NXDOMAIN
    - num.rcode.NOTIMP
    - num.rcode.REFUSED
    - num.rcode.YXDOMAIN
    - num.edns
    - num.ednserr
    - num.udp
    - num.udp6
    - num.tcp
    - num.tcp6
    - num.answer_wo_aa
    - num.rxerr
    - num.txerr
    - num.raxfr
    - num.truncated
    - num.dropped
    - zone.master
    - zone.slave
    
- nsd_thread
  - tags:
    - thread
  - fields:
    - queries

### Example Output:
```
nsd,host=localhost, num_queries=32,time_boot=340867_515436,time_elapsed=3522_901971,size_db_disk=11275648,size_db_mem=5910672,size_xfrd_mem=83979048,size_config_disk=0,size_config_mem=15600num_type_A=24,num_type_NS=1 num_opcode_QUERY=32,num_class_IN=32,num_rcode_NOERROR=16,zone_slave=8
nsd_threads,host=localhost,thread=0 num_queries=19
nsd_threads,host=localhost,thread=1 num_queries=13
```
