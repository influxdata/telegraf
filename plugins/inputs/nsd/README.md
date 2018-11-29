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

- nsd
  - fields:
    - num_queries
    - time_boot
    - time_elapsed
    - size_db_disk
    - size_db_mem
    - size_xfrd_mem
    - size_config_disk
    - size_config_mem
    - num_type_A
    - num_type_NS
    - num_type_MD
    - num_type_MF
    - num_type_CNAME
    - num_type_SOA
    - num_type_MB
    - num_type_MG
    - num_type_MR
    - num_type_NULL
    - num_type_WKS
    - num_type_PTR
    - num_type_HINFO
    - num_type_MINFO
    - num_type_MX
    - num_type_TXT
    - num_type_RP
    - num_type_AFSDB
    - num_type_X25
    - num_type_ISDN
    - num_type_RT
    - num_type_NSAP
    - num_type_SIG
    - num_type_KEY
    - num_type_PX
    - num_type_AAAA
    - num_type_LOC
    - num_type_NXT
    - num_type_SRV
    - num_type_NAPTR
    - num_type_KX
    - num_type_CERT
    - num_type_DNAME
    - num_type_OPT
    - num_type_APL
    - num_type_DS
    - num_type_SSHFP
    - num_type_IPSECKEY
    - num_type_RRSIG
    - num_type_NSEC
    - num_type_DNSKEY
    - num_type_DHCID
    - num_type_NSEC3
    - num_type_NSEC3PARAM
    - num_type_TLSA
    - num_type_SMIMEA
    - num_type_CDS
    - num_type_CDNSKEY
    - num_type_OPENPGPKEY
    - num_type_CSYNC
    - num_type_SPF
    - num_type_NID
    - num_type_L32
    - num_type_L64
    - num_type_LP
    - num_type_EUI48
    - num_type_EUI64
    - num_opcode_QUERY
    - num_class_IN
    - num_rcode_NOERROR
    - num_rcode_FORMERR
    - num_rcode_SERVFAIL
    - num_rcode_NXDOMAIN
    - num_rcode_NOTIMP
    - num_rcode_REFUSED
    - num_rcode_YXDOMAIN
    - num_edns
    - num_ednserr
    - num_udp
    - num_udp6
    - num_tcp
    - num_tcp6
    - num_answer_wo_aa
    - num_rxerr
    - num_txerr
    - num_raxfr
    - num_truncated
    - num_dropped
    - zone_master
    - zone_slave
    
- nsd_server
  - tags:
    - server
  - fields:
    - queries

### Example Output:
```
nsd,host=localhost, num_queries=32,time_boot=340867_515436,time_elapsed=3522_901971,size_db_disk=11275648,size_db_mem=5910672,size_xfrd_mem=83979048,size_config_disk=0,size_config_mem=15600num_type_A=24,num_type_NS=1 num_opcode_QUERY=32,num_class_IN=32,num_rcode_NOERROR=16,zone_slave=8
nsd_server,host=localhost,server=0 num_queries=19
nsd_server,host=localhost,server=1 num_queries=13
```
