# Unbound Input Plugin

This plugin gathers stats from [Unbound - a validating, recursive, and caching DNS resolver](https://www.unbound.net/)

### Configuration:

```toml
 # A plugin to collect stats from Unbound - a validating, recursive, and caching DNS resolver
 [[inputs.unbound]]
   ## If running as a restricted user you can prepend sudo for additional access:
   #use_sudo = false

   ## The default location of the unbound-control binary can be overridden with:
   binary = "/usr/sbin/unbound-control"

   ## The default timeout of 1s can be overriden with:
   #timeout = "1s"

   ## Use the builtin fielddrop/fieldpass telegraf filters in order to keep only specific fields
   fieldpass = ["total_*", "num_*","time_up", "mem_*"]

   ## IP of server to connect to, read from unbound conf default, optionally ':port'
   ## Will lookup IP if given a hostname
   server = "127.0.0.1:8953"
```

### Measurements & Fields:

This is the full list of stats provided by unbound-control and potentially collected by telegram
depending of your unbound configuration. Histogram related statistics will never be collected,
extended statistics can also be imported ("extended-statistics: yes" in unbound configuration).
In the output, the dots in the unbound-control stat name are replaced by underscores(see
https://www.unbound.net/documentation/unbound-control.html for details).

- unbound
    thread0_num_queries
    thread0_num_cachehits
    thread0_num_cachemiss
    thread0_num_prefetch
    thread0_num_recursivereplies
    thread0_requestlist_avg
    thread0_requestlist_max
    thread0_requestlist_overwritten
    thread0_requestlist_exceeded
    thread0_requestlist_current_all
    thread0_requestlist_current_user
    thread0_recursion_time_avg
    thread0_recursion_time_median
    total_num_queries
    total_num_cachehits
    total_num_cachemiss
    total_num_prefetch
    total_num_recursivereplies
    total_requestlist_avg
    total_requestlist_max
    total_requestlist_overwritten
    total_requestlist_exceeded
    total_requestlist_current_all
    total_requestlist_current_user
    total_recursion_time_avg
    total_recursion_time_median
    time_now
    time_up
    time_elapsed
    mem_total_sbrk
    mem_cache_rrset
    mem_cache_message
    mem_mod_iterator
    mem_mod_validator
    num_query_type_A
    num_query_type_PTR
    num_query_type_TXT
    num_query_type_AAAA
    num_query_type_SRV
    num_query_type_ANY
    num_query_class_IN
    num_query_opcode_QUERY
    num_query_tcp
    num_query_ipv6
    num_query_flags_QR
    num_query_flags_AA
    num_query_flags_TC
    num_query_flags_RD
    num_query_flags_RA
    num_query_flags_Z
    num_query_flags_AD
    num_query_flags_CD
    num_query_edns_present
    num_query_edns_DO
    num_answer_rcode_NOERROR
    num_answer_rcode_SERVFAIL
    num_answer_rcode_NXDOMAIN
    num_answer_rcode_nodata
    num_answer_secure
    num_answer_bogus
    num_rrset_bogus
    unwanted_queries
    unwanted_replies

### Permissions:

It's important to note that this plugin references unbound-control, which may require additional permissions to execute successfully.
Depending on the user/group permissions of the telegraf user executing this plugin, you may need to alter the group membership, set facls, or use sudo.

**Group membership (Recommended)**:
```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G unbound telegraf

$ groups telegraf
telegraf : telegraf unbound
```

**Sudo privileges**:
If you use this method, you will need the following in your telegraf config:
```toml
[[inputs.unbound]]
  use_sudo = true
```

You will also need to update your sudoers file:
```bash
$ visudo
# Add the following line:
telegraf ALL=(ALL) NOPASSWD: /usr/sbin/unbound-control
```

Please use the solution you see as most appropriate.

### Example Output:

```
 telegraf --config etc/telegraf.conf --input-filter unbound --test
* Plugin: inputs.unbound, Collection 1
> unbound,host=localhost total_num_cachehits=0,total_num_prefetch=0,total_requestlist_avg=0,total_requestlist_max=0,total_recursion_time_median=0,total_num_queries=0,total_requestlist_overwritten=0,total_requestlist_current_all=0,time_up=159185.583967,total_num_recursivereplies=0,total_requestlist_exceeded=0,total_requestlist_current_user=0,total_recursion_time_avg=0,total_tcpusage=0,total_num_cachemiss=0 1510130793000000000

```
