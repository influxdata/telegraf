# Varnish Input Plugin

This plugin gathers stats from [Unbound - a validating, recursive, and caching DNS resolver](https://www.unbound.net/)

### Configuration:

```toml
 # A plugin to collect stats from Unbound - a validating, recursive, and caching DNS resolver
 [[inputs.varnish]]
   ## If running as a restricted user you can prepend sudo for additional access:
   #use_sudo = false

   ## The default location of the varnishstat binary can be overridden with:
   binary = "/usr/sbin/unbound-control"

   ## By default, telegraf gathers stats for 3 metric points.
   ## Setting stats will override the defaults shown below.
   ## stats may also be set to ["all"], which will collect all stats
   stats = ["total.*", "num.*","time.up", "mem.*"]
```

### Measurements & Fields:

This is the full list of stats provided by unbound. Stats will be grouped by their prefix (eg thread0,
total, etc). In the output, the prefix will be used as a tag, and removed from field names. See
https://www.unbound.net/documentation/unbound-control.html for details.

- unbound
    thread0.num.queries
    thread0.num.cachehits
    thread0.num.cachemiss
    thread0.num.prefetch
    thread0.num.recursivereplies
    thread0.requestlist.avg
    thread0.requestlist.max
    thread0.requestlist.overwritten
    thread0.requestlist.exceeded
    thread0.requestlist.current.all
    thread0.requestlist.current.user
    thread0.recursion.time.avg
    thread0.recursion.time.median
    total.num.queries
    total.num.cachehits
    total.num.cachemiss
    total.num.prefetch
    total.num.recursivereplies
    total.requestlist.avg
    total.requestlist.max
    total.requestlist.overwritten
    total.requestlist.exceeded
    total.requestlist.current.all
    total.requestlist.current.user
    total.recursion.time.avg
    total.recursion.time.median
    time.now
    time.up
    time.elapsed
    mem.total.sbrk
    mem.cache.rrset
    mem.cache.message
    mem.mod.iterator
    mem.mod.validator
    histogram.000000.000000.to.000000.000001
    histogram.000000.000001.to.000000.000002
    histogram.000000.000002.to.000000.000004
    histogram.000000.000004.to.000000.000008
    histogram.000000.000008.to.000000.000016
    histogram.000000.000016.to.000000.000032
    histogram.000000.000032.to.000000.000064
    histogram.000000.000064.to.000000.000128
    histogram.000000.000128.to.000000.000256
    histogram.000000.000256.to.000000.000512
    histogram.000000.000512.to.000000.001024
    histogram.000000.001024.to.000000.002048
    histogram.000000.002048.to.000000.004096
    histogram.000000.004096.to.000000.008192
    histogram.000000.008192.to.000000.016384
    histogram.000000.016384.to.000000.032768
    histogram.000000.032768.to.000000.065536
    histogram.000000.065536.to.000000.131072
    histogram.000000.131072.to.000000.262144
    histogram.000000.262144.to.000000.524288
    histogram.000000.524288.to.000001.000000
    histogram.000001.000000.to.000002.000000
    histogram.000002.000000.to.000004.000000
    histogram.000004.000000.to.000008.000000
    histogram.000008.000000.to.000016.000000
    histogram.000016.000000.to.000032.000000
    histogram.000032.000000.to.000064.000000
    histogram.000064.000000.to.000128.000000
    histogram.000128.000000.to.000256.000000
    histogram.000256.000000.to.000512.000000
    histogram.000512.000000.to.001024.000000
    histogram.001024.000000.to.002048.000000
    histogram.002048.000000.to.004096.000000
    histogram.004096.000000.to.008192.000000
    histogram.008192.000000.to.016384.000000
    histogram.016384.000000.to.032768.000000
    histogram.032768.000000.to.065536.000000
    histogram.065536.000000.to.131072.000000
    histogram.131072.000000.to.262144.000000
    histogram.262144.000000.to.524288.000000
    num.query.type.A
    num.query.type.PTR
    num.query.type.TXT
    num.query.type.AAAA
    num.query.type.SRV
    num.query.type.ANY
    num.query.class.IN
    num.query.opcode.QUERY
    num.query.tcp
    num.query.ipv6
    num.query.flags.QR
    num.query.flags.AA
    num.query.flags.TC
    num.query.flags.RD
    num.query.flags.RA
    num.query.flags.Z
    num.query.flags.AD
    num.query.flags.CD
    num.query.edns.present
    num.query.edns.DO
    num.answer.rcode.NOERROR
    num.answer.rcode.SERVFAIL
    num.answer.rcode.NXDOMAIN
    num.answer.rcode.nodata
    num.answer.secure
    num.answer.bogus
    num.rrset.bogus
    unwanted.queries
    unwanted.replies

### Tags:

As indicated above, the  prefix of a unbound stat will be used as it's 'section' tag. So section tag may have one of
the following values:
- section:
      - thread0
      - total
      - time
      - mem
      - histogram
      - num
      - unwanted

### Permissions:

It's important to note that this plugin references unbound-control, which may require additional permissions to execute successfully.
Depending on the user/group permissions of the telegraf user executing this plugin, you may need to alter the group membership, set facls, or use sudo.

**Group membership (Recommended)**:
```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G unbound telegraf

$ groups telegraf
telegraf : telegraf varnish
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
> unbound,section=total,host=laptop-aromeyer num.cachemiss=0,requestlist.current.all=0,num.cachehits=0,requestlist.overwritten=0,requestlist.max=0,num.recursivereplies=0,requestlist.avg=0,recursion.time.avg=0,recursion.time.median=0,num.prefetch=0,requestlist.exceeded=0,requestlist.current.user=0,tcpusage=0,num.queries=0 1509977403000000000
> unbound,section=time,host=laptop-aromeyer up=5794.844261,elapsed=12.484727,now=1509977402.617432 1509977403000000000

```
