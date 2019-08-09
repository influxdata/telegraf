# Suricata plugin for Telegraf

This plugin reports internal performance counters of the Suricata IDS/IPS
engine, such as captured traffic volume, memory usage, uptime, flow counters,
and much more. It provides a socket for the Suricata log output to write JSON
stats output to, and processes the incoming data to fit Telegraf's format.

### Configuration:

```toml
[[input.suricata]]
  ## Data sink for Suricata stats log.
  # This is expected to be a filename of a
  # unix socket to be created for listening.
  # Will be overwritten if a socket or file
  # with that name already exists.
  source = "/tmp/suricata-stats.sock"
```

### Measurements & Fields:

Fields in the 'suricata' measurement follow the JSON format used by Suricata's
stats output.
See http://suricata.readthedocs.io/en/latest/performance/statistics.html for
more information.

All fields are numeric.

- suricata
    - app_layer.flow.dcerpc_udp
    - app_layer.flow.dns_tcp
    - app_layer.flow.dns_udp
    - app_layer.flow.enip_udp
    - app_layer.flow.failed_tcp
    - app_layer.flow.failed_udp
    - app_layer.flow.http
    - app_layer.flow.ssh
    - app_layer.flow.tls
    - app_layer.tx.dns_tcp
    - app_layer.tx.dns_udp
    - app_layer.tx.enip_udp
    - app_layer.tx.http
    - app_layer.tx.smtp
    - capture.kernel_drops
    - capture.kernel_packets
    - decoder.avg_pkt_size
    - decoder.bytes
    - decoder.ethernet
    - decoder.gre
    - decoder.icmpv4
    - decoder.icmpv4.ipv4_unknown_ver
    - decoder.icmpv6
    - decoder.invalid
    - decoder.ipv4
    - decoder.ipv6
    - decoder.max_pkt_size
    - decoder.pkts
    - decoder.tcp
    - decoder.tcp.hlen_too_small
    - decoder.tcp.invalid_optlen
    - decoder.teredo
    - decoder.udp
    - decoder.vlan
    - detect.alert
    - dns.memcap_global
    - dns.memuse
    - flow.memuse
    - flow_mgr.closed_pruned
    - flow_mgr.est_pruned
    - flow_mgr.flows_checked
    - flow_mgr.flows_notimeout
    - flow_mgr.flows_removed
    - flow_mgr.flows_timeout
    - flow_mgr.flows_timeout_inuse
    - flow_mgr.new_pruned
    - flow_mgr.rows_checked
    - flow_mgr.rows_empty
    - flow_mgr.rows_maxlen
    - flow_mgr.rows_skipped
    - flow.spare
    - flow.tcp_reuse
    - http.memuse
    - tcp.memuse
    - tcp.pseudo
    - tcp.reassembly_gap
    - tcp.reassembly_memuse
    - tcp.rst
    - tcp.sessions
    - tcp.syn
    - tcp.synack
    - ...


### Tags:

The `suricata` measurement has the following tags:

- thread: `Global` for global statistics, thread IDs (e.g. `W#03-enp0s31f6`) for thread-specific statistics, `total` for a summary across all threads.

## Suricata configuration

Suricata needs to deliver the 'stats' event type to a given unix socket for
this plugin to pick up. This can be done, for example, by creating an additional
output in the Suricata configuration file:

```yaml
- eve-log:
    enabled: yes
    filetype: unix_stream
    filename: /tmp/suricata-stats.sock
    types:
      - stats:
         threads: yes
```

## Example Output:

```text
suricata,host=BLN02NB0124,thread=Global tcp.reassembly_memuse=12332832,dns.memuse=0,dns.memcap_state=0,flow.memuse=7074304,dns.memcap_global=0,http.memuse=0,http.memcap=0,tcp.memuse=1638400 1501764050000000000
suricata,thread=W#03-enp0s31f6,host=BLN02NB0124 app_layer.flow.failed_tcp=0,decoder.invalid=0,decoder.ipv4=0,capture.kernel_packets=0,defrag.ipv6.reassembled=0,decoder.icmpv6=0,app_layer.flow.ftp=0,app_layer.tx.dns_udp=0,decoder.null=0,decoder.pppoe=0,tcp.rst=0,decoder.erspan=0,decoder.gre=0,tcp.reassembly_gap=0,app_layer.tx.http=0,app_layer.flow.smtp=0,decoder.bytes=0,app_layer.tx.smtp=0,app_layer.flow.ssh=0,decoder.teredo=0,decoder.avg_pkt_size=0,tcp.synack=0,app_layer.flow.msn=0,app_layer.tx.dns_tcp=0,tcp.segment_memcap_drop=0,defrag.max_frag_hits=0,capture.kernel_drops=0,defrag.ipv4.timeouts=0,decoder.ppp=0,decoder.sctp=0,decoder.ipv6_in_ipv6=0,flow.memcap=0,tcp.pseudo_failed=0,decoder.pkts=0,defrag.ipv6.fragments=0,tcp.syn=0,app_layer.flow.failed_udp=0,decoder.vlan_qinq=0,defrag.ipv4.fragments=0,app_layer.flow.dns_tcp=0,tcp.pseudo=0,decoder.mpls=0,tcp.sessions=0,decoder.raw=0,decoder.ipv6=0,decoder.ltnull.unsupported_type=0,app_layer.flow.dns_udp=0,tcp.no_flow=0,decoder.sll=0,decoder.ltnull.pkt_too_small=0,app_layer.flow.smb=0,decoder.ipv4_in_ipv6=0,app_layer.tx.tls=0,app_layer.flow.dcerpc_udp=0,decoder.max_pkt_size=0,app_layer.flow.imap=0,tcp.invalid_checksum=0,app_layer.flow.http=0,decoder.tcp=0,decoder.vlan=0,app_layer.flow.tls=0,decoder.ethernet=0,tcp.ssn_memcap_drop=0,app_layer.flow.dcerpc_tcp=0,decoder.dce.pkt_too_small=0,defrag.ipv4.reassembled=0,tcp.stream_depth_reached=0,decoder.ipraw.invalid_ip_version=0,decoder.icmpv4=0,detect.alert=0,defrag.ipv6.timeouts=0,decoder.udp=0 1501764050000000000
suricata,thread=W#04-enp0s31f6,host=BLN02NB0124 app_layer.flow.failed_udp=0,decoder.avg_pkt_size=0,decoder.pkts=0,app_layer.flow.ftp=0,app_layer.flow.http=0,tcp.no_flow=0,defrag.ipv4.timeouts=0,tcp.rst=0,decoder.sctp=0,decoder.vlan=0,defrag.max_frag_hits=0,decoder.icmpv4=0,decoder.ltnull.unsupported_type=0,tcp.reassembly_gap=0,tcp.stream_depth_reached=0,decoder.null=0,app_layer.flow.tls=0,defrag.ipv4.reassembled=0,decoder.vlan_qinq=0,capture.kernel_drops=0,decoder.ethernet=0,defrag.ipv6.fragments=0,capture.kernel_packets=0,decoder.erspan=0,decoder.max_pkt_size=0,decoder.udp=0,detect.alert=0,decoder.ipraw.invalid_ip_version=0,app_layer.tx.tls=0,decoder.bytes=0,app_layer.flow.failed_tcp=0,decoder.ipv6=0,decoder.ppp=0,app_layer.flow.dns_tcp=0,app_layer.flow.imap=0,decoder.tcp=0,decoder.ipv6_in_ipv6=0,decoder.sll=0,tcp.synack=0,tcp.segment_memcap_drop=0,tcp.sessions=0,decoder.ltnull.pkt_too_small=0,decoder.pppoe=0,decoder.mpls=0,app_layer.tx.smtp=0,app_layer.tx.dns_udp=0,tcp.pseudo_failed=0,app_l2017-08-03T12:41:00Z D! Output [file] wrote batch of 61 metrics in 217.873µs
ayer.flow.msn=0,decoder.gre=0,defrag.ipv6.reassembled=0,tcp.syn=0,app_layer.flow.dns_udp=0,tcp.invalid_checksum=0,tcp.pseudo=0,decoder.dce.pkt_too_small=0,decoder.icmpv6=0,app_layer.flow.smtp=0,app_layer.flow.smb=0,decoder.raw=0,decoder.teredo=0,app_layer.flow.dcerpc_udp=0,decoder.ipv4=0,app_layer.flow.dcerpc_tcp=0,app_layer.tx.dns_tcp=0,flow.memcap=0,defrag.ipv6.timeouts=0,decoder.ipv4_in_ipv6=0,defrag.ipv4.fragments=0,tcp.ssn_memcap_drop=0,app_layer.tx.http=0,app_layer.flow.ssh=0,decoder.invalid=0 1501764050000000000
suricata,thread=FM#01,host=BLN02NB0124 flow.emerg_mode_entered=0,flow_mgr.flows_timeout=0,flow_mgr.flows_removed=0,flow_mgr.new_pruned=0,flow_mgr.rows_empty=0,flow.emerg_mode_over=0,flow_mgr.rows_skipped=65536,flow_mgr.rows_checked=65536,flow_mgr.flows_checked=0,flow_mgr.rows_busy=0,flow.spare=10000,flow.tcp_reuse=0,flow_mgr.flows_notimeout=0,flow_mgr.rows_maxlen=0,flow_mgr.bypassed_pruned=0,flow_mgr.flows_timeout_inuse=0,flow_mgr.est_pruned=0,flow_mgr.closed_pruned=0 1501764050000000000
suricata,thread=total,host=BLN02NB0124 app_layer.flow.failed_tcp=0,flow.spare=10000,app_layer.flow.ftp=0,app_layer.tx.tls=0,decoder.null=0,tcp.ssn_memcap_drop=0,capture.kernel_drops=0,decoder.tcp=0,event_type="stats",timestamp="2017-08-03T14:40:50.000118+0200",app_layer.tx.dns_tcp=0,flow.tcp_reuse=0,tcp.pseudo_failed=0,decoder.dce.pkt_too_small=0,decoder.teredo=0,decoder.ipv6=0,decoder.ipraw.invalid_ip_version=0,app_layer.flow.http=0,defrag.ipv4.timeouts=0,decoder.ipv4=0,defrag.ipv4.reassembled=0,uptime=3763,decoder.vlan=0,dns.memcap_state=0,decoder.udp=0,flow_mgr.rows_skipped=65536,decoder.sctp=0,decoder.icmpv4=0,decoder.ltnull.unsupported_type=0,decoder.pkts=0,flow_mgr.rows_busy=0,decoder.invalid=0,flow_mgr.est_pruned=0,flow_mgr.rows_checked=65536,dns.memuse=0,app_layer.flow.smb=0,decoder.ethernet=0,app_layer.flow.dcerpc_udp=0,decoder.pppoe=0,app_layer.flow.dns_tcp=0,flow_mgr.flows_checked=0,detect.alert=0,app_layer.flow.msn=0,decoder.gre=0,capture.kernel_packets=0,http.memuse=0,flow.memcap=0,tcp.pseudo=0,defrag.ipv6.timeouts=0,flow.memuse=7074304,flow_mgr.flows_timeout_inuse=0,tcp.reassembly_gap=0,defrag.ipv4.fragments=0,app_layer.flow.tls=0,decoder.icmpv6=0,app_layer.flow.failed_udp=0,tcp.rst=0,decoder.vlan_qinq=0,tcp.invalid_checksum=0,decoder.max_pkt_size=0,dns.memcap_global=0,app_layer.tx.http=0,decoder.erspan=0,tcp.synack=0,app_layer.flow.imap=0,flow_mgr.flows_timeout=0,tcp.no_flow=0,flow_mgr.flows_notimeout=0,flow_mgr.new_pruned=0,flow_mgr.rows_empty=0,flow_mgr.bypassed_pruned=0,http.memcap=0,app_layer.tx.dns_udp=0,tcp.syn=0,tcp.memuse=1638400,app_layer.flow.dns_udp=0,decoder.ltnull.pkt_too_small=0,tcp.stream_depth_reached=0,app_layer.flow.ssh=0,defrag.ipv6.reassembled=0,tcp.reassembly_memuse=12332832,decoder.sll=0,flow_mgr.flows_removed=0,tcp.segment_memcap_drop=0,app_layer.flow.dcerpc_tcp=0,defrag.max_frag_hits=0,app_layer.flow.smtp=0,defrag.ipv6.fragments=0,flow_mgr.rows_maxlen=0,decoder.raw=0,decoder.bytes=0,decoder.avg_pkt_size=0,tcp.sessions=0,decoder.ipv4_in_ipv6=0,flow.emerg_mode_over=0,flow.emerg_mode_entered=0,app_layer.tx.smtp=0,decoder.ppp=0,decoder.mpls=0,flow_mgr.closed_pruned=0,decoder.ipv6_in_ipv6=0 1501764050000000000
suricata,thread=W#02-enp0s31f6,host=BLN02NB0124 decoder.ethernet=0,app_layer.flow.imap=0,app_layer.flow.msn=0,defrag.ipv4.fragments=0,tcp.pseudo=0,app_layer.flow.dcerpc_udp=0,decoder.gre=0,decoder.ppp=0,tcp.sessions=0,capture.kernel_packets=0,defrag.ipv6.reassembled=0,decoder.bytes=0,decoder.max_pkt_size=0,app_layer.flow.smb=0,defrag.ipv6.fragments=0,decoder.ipv6=0,app_layer.flow.dcerpc_tcp=0,decoder.pppoe=0,decoder.vlan=0,tcp.no_flow=0,decoder.teredo=0,app_layer.flow.smtp=0,decoder.ltnull.pkt_too_small=0,decoder.vlan_qinq=0,decoder.dce.pkt_too_small=0,decoder.avg_pkt_size=0,decoder.icmpv4=0,decoder.tcp=0,tcp.synack=0,decoder.ipv4_in_ipv6=0,defrag.max_frag_hits=0,decoder.mpls=0,app_layer.tx.smtp=0,defrag.ipv4.timeouts=0,decoder.ltnull.unsupported_type=0,app_layer.tx.dns_tcp=0,tcp.stream_depth_reached=0,decoder.udp=0,app_layer.flow.tls=0,decoder.pkts=0,tcp.syn=0,decoder.ipv6_in_ipv6=0,app_layer.flow.dns_tcp=0,app_layer.flow.failed_tcp=0,defrag.ipv6.timeouts=0,decoder.invalid=0,app_layer.flow.ftp=0,decoder.ipv4=0,app_layer.flow.dns_udp=0,flow.memcap=0,decoder.raw=0,app_layer.flow.http=0,tcp.rst=0,app_layer.tx.dns_udp=0,tcp.ssn_memcap_drop=0,decoder.sctp=0,decoder.ipraw.invalid_ip_version=0,decoder.null=0,app_layer.tx.tls=0,detect.alert=0,decoder.erspan=0,app_layer.tx.http=0,decoder.icmpv6=0,tcp.invalid_checksum=0,app_layer.flow.failed_udp=0,tcp.reassembly_gap=0,defrag.ipv4.reassembled=0,tcp.pseudo_failed=0,decoder.sll=0,tcp.segment_memcap_drop=0,capture.kernel_drops=0,app_layer.flow.ssh=0 1501764050000000000
suricata,thread=W#01-enp0s31f6,host=BLN02NB0124 decoder.gre=0,decoder.max_pkt_size=0,app_layer.flow.dcerpc_udp=0,tcp.invalid_checksum=0,decoder.pkts=0,defrag.ipv4.timeouts=0,app_layer.tx.dns_tcp=0,app_layer.flow.msn=0,decoder.ltnull.unsupported_type=0,tcp.ssn_memcap_drop=0,app_layer.flow.tls=0,app_layer.tx.http=0,app_layer.flow.http=0,decoder.ipraw.invalid_ip_version=0,app_layer.flow.dcerpc_tcp=0,app_layer.tx.tls=0,decoder.mpls=0,tcp.no_flow=0,decoder.vlan=0,decoder.sll=0,app_layer.flow.dns_tcp=0,decoder.erspan=0,decoder.ipv6_in_ipv6=0,detect.alert=0,tcp.sessions=0,decoder.pppoe=0,decoder.ipv6=0,defrag.ipv6.reassembled=0,decoder.vlan_qinq=0,app_layer.flow.dns_udp=0,defrag.ipv4.reassembled=0,decoder.udp=0,app_layer.tx.smtp=0,tcp.pseudo=0,capture.kernel_drops=0,app_layer.flow.smb=0,capture.kernel_packets=0,decoder.tcp=0,decoder.avg_pkt_size=0,app_layer.flow.failed_tcp=0,decoder.icmpv4=0,tcp.stream_depth_reached=0,decoder.icmpv6=0,defrag.ipv6.timeouts=0,decoder.ltnull.pkt_too_small=0,decoder.sctp=0,defrag.ipv4.fragments=0,app_layer.flow.ftp=0,tcp.synack=0,tcp.reassembly_gap=0,tcp.rst=0,decoder.ethernet=0,app_layer.flow.ssh=0,defrag.max_frag_hits=0,defrag.ipv6.fragments=0,decoder.teredo=0,app_layer.flow.smtp=0,tcp.segment_memcap_drop=0,decoder.ipv4=0,decoder.ipv4_in_ipv6=0,decoder.dce.pkt_too_small=0,app_layer.flow.imap=0,app_layer.flow.failed_udp=0,tcp.pseudo_failed=0,decoder.invalid=0,decoder.null=0,tcp.syn=0,decoder.raw=0,app_layer.tx.dns_udp=0,flow.memcap=0,decoder.bytes=0,decoder.ppp=0 1501764050000000000
suricata,thread=W#01-enp0s31f6,host=BLN02NB0124 app_layer.tx.dns_tcp=0,app_layer.tx.tls=0,tcp.reassembly_gap=0,app_layer.flow.dcerpc_udp=0,app_layer.tx.dns_udp=0,tcp.invalid_checksum=0,app_layer.flow.tls=0,decoder.ppp=0,decoder.teredo=0,defrag.ipv6.reassembled=0,decoder.ethernet=0,tcp.no_flow=0,app_layer.tx.smtp=0,app_layer.flow.dcerpc_tcp=0,decoder.icmpv6=0,decoder.sctp=0,decoder.ipv4=0,app_layer.flow.failed_tcp=0,app_layer.flow.http=0,decoder.ipv6=0,defrag.ipv4.reassembled=0,decoder.mpls=0,flow.memcap=0,decoder.max_pkt_size=0,decoder.raw=0,app_layer.tx.http=0,detect.alert=0,tcp.syn=0,decoder.avg_pkt_size=0,app_layer.flow.imap=0,tcp.sessions=0,decoder.tcp=0,app_layer.flow.smtp=0,decoder.ipraw.invalid_ip_version=0,tcp.stream_depth_reached=0,app_layer.flow.ftp=0,tcp.pseudo=0,tcp.synack=0,app_layer.flow.msn=0,app_layer.flow.failed_udp=0,decoder.invalid=0,decoder.sll=0,tcp.rst=0,decoder.pkts=0,defrag.ipv4.fragments=0,decoder.vlan_qinq=0,decoder.bytes=0,app_layer.flow.dns_udp=0,tcp.pseudo_failed=0,capture.kernel_packets=0,app_layer.flow.dns_tcp=0,decoder.udp=0,defrag.ipv4.timeouts=0,decoder.ipv4_in_ipv6=0,defrag.max_frag_hits=0,decoder.ipv6_in_ipv6=0,decoder.dce.pkt_too_small=0,capture.kernel_drops=0,app_layer.flow.smb=0,decoder.ltnull.pkt_too_small=0,tcp.ssn_memcap_drop=0,decoder.erspan=0,defrag.ipv6.timeouts=0,decoder.vlan=0,app_layer.flow.ssh=0,decoder.icmpv4=0,tcp.segment_memcap_drop=0,decoder.gre=0,decoder.ltnull.unsupported_type=0,decoder.null=0,decoder.pppoe=0,defrag.ipv6.fragments=0 1501764060000000000
suricata,thread=FM#01,host=BLN02NB0124 flow_mgr.flows_removed=0,flow_mgr.flows_notimeout=0,flow_mgr.flows_checked=0,flow_mgr.closed_pruned=0,flow_mgr.rows_maxlen=0,flow_mgr.rows_empty=0,flow.emerg_mode_entered=0,flow_mgr.flows_timeout_inuse=0,flow_mgr.rows_checked=65536,flow_mgr.flows_timeout=0,flow_mgr.new_pruned=0,flow_mgr.bypassed_pruned=0,flow.emerg_mode_over=0,flow.tcp_reuse=0,flow_mgr.est_pruned=0,flow_mgr.rows_busy=0,flow_mgr.rows_skipped=65536,flow.spare=10000 1501764060000000000
suricata,thread=W#02-enp0s31f6,host=BLN02NB0124 app_layer.flow.tls=0,decoder.gre=0,app_layer.tx.smtp=0,decoder.ethernet=0,decoder.icmpv4=0,decoder.pppoe=0,decoder.erspan=0,decoder.dce.pkt_too_small=0,tcp.reassembly_gap=0,app_layer.flow.smtp=0,tcp.pseudo_failed=0,decoder.invalid=0,decoder.ltnull.unsupported_type=0,app_layer.flow.ftp=0,defrag.ipv4.timeouts=0,decoder.ltnull.pkt_too_small=0,decoder.ipv6=0,decoder.null=0,tcp.syn=0,decoder.avg_pkt_size=0,app_layer.flow.dns_udp=0,capture.kernel_packets=0,decoder.pkts=0,app_layer.flow.ssh=0,decoder.ipraw.invalid_ip_version=0,decoder.ipv4=0,decoder.ipv6_in_ipv6=0,app_layer.tx.dns_tcp=0,app_layer.tx.tls=0,defrag.ipv4.reassembled=0,capture.kernel_drops=0,tcp.no_flow=0,app_layer.flow.dcerpc_tcp=0,tcp.sessions=0,decoder.tcp=0,app_layer.flow.dns_tcp=0,decoder.bytes=0,app_layer.flow.dcerpc_udp=0,tcp.synack=0,decoder.mpls=0,decoder.ipv4_in_ipv6=0,decoder.udp=0,defrag.ipv6.fragments=0,decoder.raw=0,defrag.ipv4.fragments=0,app_layer.flow.failed_udp=0,app_layer.flow.failed_tcp=0,tcp.pseudo=0,app_layer.flow.imap=0,app_layer.flow.http=0,decoder.vlan=0,tcp.rst=0,decoder.sll=0,app_layer.flow.smb=0,decoder.max_pkt_size=0,defrag.max_frag_hits=0,detect.alert=0,app_layer.tx.http=0,flow.memcap=0,tcp.invalid_checksum=0,defrag.ipv6.timeouts=0,decoder.icmpv6=0,tcp.ssn_memcap_drop=0,decoder.teredo=0,tcp.segment_memcap_drop=0,tcp.stream_depth_reached=0,app_layer.tx.dns_udp=0,decoder.ppp=0,defrag.ipv6.reassembled=0,decoder.sctp=0,decoder.vlan_qinq=0,app_layer.flow.msn=0 1501764060000000000
suricata,thread=W#04-enp0s31f6,host=BLN02NB0124 decoder.udp=0,tcp.invalid_checksum=0,decoder.ppp=0,decoder.sll=0,app_layer.flow.msn=0,app_layer.flow.dns_udp=0,tcp.pseudo=0,decoder.ipv6_in_ipv6=0,decoder.pppoe=0,defrag.max_frag_hits=0,detect.alert=0,decoder.mpls=0,tcp.synack=0,decoder.ltnull.unsupported_type=0,decoder.ipraw.invalid_ip_version=0,tcp.segment_memcap_drop=0,decoder.dce.pkt_too_small=0,app_layer.tx.dns_tcp=0,app_layer.flow.dcerpc_tcp=0,capture.kernel_packets=0,capture.kernel_drops=0,decoder.erspan=0,decoder.teredo=0,app_layer.flow.dns_tcp=0,defrag.ipv4.fragments=0,decoder.vlan_qinq=0,app_layer.flow.smb=0,app_layer.flow.failed_udp=0,app_layer.tx.tls=0,tcp.reassembly_gap=0,decoder.bytes=0,decoder.tcp=0,tcp.stream_depth_reached=0,decoder.ipv4=0,tcp.ssn_memcap_drop=0,decoder.icmpv6=0,app_layer.flow.ftp=0,app_layer.flow.http=0,tcp.rst=0,decoder.ipv4_in_ipv6=0,app_layer.flow.tls=0,defrag.ipv4.timeouts=0,app_layer.flow.failed_tcp=0,app_layer.flow.smtp=0,tcp.pseudo_failed=0,defrag.ipv6.fragments=0,app_layer.tx.smtp=0,tcp.sessions=0,decoder.max_pkt_size=0,decoder.icmpv4=0,defrag.ipv4.reassembled=0,app_layer.tx.http=0,decoder.raw=0,defrag.ipv6.timeouts=0,tcp.syn=0,flow.memcap=0,decoder.pkts=0,defrag.ipv6.reassembled=0,decoder.ethernet=0,decoder.gre=0,app_layer.tx.dns_udp=0,tcp.no_flow=0,app_layer.flow.dcerpc_udp=0,decoder.ltnull.pkt_too_small=0,decoder.vlan=0,decoder.null=0,decoder.sctp=0,decoder.invalid=0,decoder.ipv6=0,decoder.avg_pkt_size=0,app_layer.flow.imap=0,app_layer.flow.ssh=0 1501764060000000000
suricata,thread=Global,host=BLN02NB0124 tcp.reassembly_memuse=12332832,tcp.memuse=1638400,dns.memuse=0,flow.memuse=7074304,dns.memcap_state=0,dns.memcap_global=0,http.memcap=0,http.memuse=0 1501764060000000000
suricata,thread=W#03-enp0s31f6,host=BLN02NB0124 app_layer.tx.smtp=0,tcp.pseudo=0,decoder.icmpv6=0,decoder.ipv6_in_ipv6=0,defrag.ipv6.timeouts=0,defrag.ipv4.reassembled=0,tcp.synack=0,app_layer.flow.dcerpc_udp=0,app_layer.flow.dns_tcp=0,tcp.invalid_checksum=0,app_layer.flow.smb=0,decoder.udp=0,app_layer.flow.msn=0,decoder.ltnull.pkt_too_small=0,decoder.sll=0,app_layer.flow.http=0,decoder.bytes=0,defrag.ipv4.fragments=0,decoder.teredo=0,decoder.erspan=0,defrag.max_frag_hits=0,decoder.vlan_qinq=0,tcp.syn=0,decoder.ethernet=0,app_layer.flow.dcerpc_tcp=0,decoder.ltnull.unsupported_type=0,tcp.stream_depth_reached=0,app_layer.tx.dns_udp=0,app_layer.flow.ssh=0,capture.kernel_drops=0,defrag.ipv6.reassembled=0,decoder.raw=0,decoder.gre=0,defrag.ipv6.fragments=0,tcp.reassembly_gap=0,decoder.pppoe=0,app_layer.flow.dns_udp=0,decoder.ipv6=0,tcp.sessions=0,app_layer.tx.http=0,decoder.sctp=0,app_layer.tx.dns_tcp=0,tcp.no_flow=0,decoder.pkts=0,decoder.icmpv4=0,flow.memcap=0,app_layer.flow.smtp=0,tcp.ssn_memcap_drop=0,decoder.max_pkt_size=0,tcp.segment_memcap_drop=0,decoder.dce.pkt_too_small=0,app_layer.flow.failed_tcp=0,decoder.mpls=0,decoder.invalid=0,app_layer.flow.ftp=0,decoder.vlan=0,defrag.ipv4.timeouts=0,decoder.ipraw.invalid_ip_version=0,capture.kernel_packets=0,tcp.rst=0,decoder.avg_pkt_size=0,app_layer.tx.tls=0,decoder.ipv4=0,decoder.null=0,decoder.tcp=0,detect.alert=0,decoder.ppp=0,app_layer.flow.failed_udp=0,app_layer.flow.tls=0,decoder.ipv4_in_ipv6=0,app_layer.flow.imap=0,tcp.pseudo_failed=0 1501764060000000000
suricata,thread=total,host=BLN02NB0124 decoder.teredo=0,flow.emerg_mode_entered=0,dns.memcap_global=0,tcp.pseudo=0,uptime=3772,tcp.ssn_memcap_drop=0,app_layer.flow.tls=0,defrag.ipv6.fragments=0,tcp.memuse=1638400,tcp.pseudo_failed=0,app_layer.flow.dcerpc_tcp=0,decoder.max_pkt_size=0,tcp.sessions=0,decoder.raw=0,app_layer.flow.failed_tcp=0,flow_mgr.flows_timeout=0,decoder.ipv4=0,flow.spare=10000,decoder.vlan=0,flow_mgr.flows_notimeout=0,decoder.erspan=0,capture.kernel_packets=0,decoder.avg_pkt_size=0,capture.kernel_drops=0,detect.alert=0,app_layer.tx.tls=0,tcp.synack=0,flow_mgr.flows_timeout_inuse=0,dns.memcap_state=0,app_layer.flow.dns_udp=0,flow_mgr.rows_empty=0,flow.memuse=7074304,decoder.gre=0,app_layer.flow.smb=0,tcp.rst=0,decoder.sll=0,decoder.null=0,app_layer.tx.dns_udp=0,flow_mgr.rows_checked=65536,app_layer.flow.imap=0,tcp.reassembly_gap=0,decoder.ipv6=0,decoder.mpls=0,event_type="stats",decoder.dce.pkt_too_small=0,app_layer.flow.msn=0,http.memcap=0,defrag.ipv6.timeouts=0,app_layer.tx.dns_tcp=0,defrag.ipv4.fragments=0,defrag.ipv6.reassembled=0,decoder.pppoe=0,flow_mgr.closed_pruned=0,timestamp="2017-08-03T14:40:59.000486+0200",flow_mgr.flows_removed=0,flow_mgr.rows_skipped=65536,app_layer.flow.ftp=0,app_layer.flow.smtp=0,defrag.ipv4.reassembled=0,app_layer.flow.ssh=0,tcp.segment_memcap_drop=0,decoder.pkts=0,decoder.vlan_qinq=0,decoder.ethernet=0,flow_mgr.new_pruned=0,tcp.stream_depth_reached=0,flow.tcp_reuse=0,flow_mgr.flows_checked=0,flow_mgr.rows_maxlen=0,decoder.ipv4_in_ipv6=0,app_layer.flow.failed_udp=0,decoder.icmpv6=0,decoder.ipv6_in_ipv6=0,http.memuse=0,decoder.ltnull.unsupported_type=0,decoder.icmpv4=0,app_layer.tx.http=0,decoder.tcp=0,tcp.syn=0,decoder.sctp=0,app_layer.flow.http=0,decoder.bytes=0,decoder.ppp=0,app_layer.flow.dcerpc_udp=0,tcp.invalid_checksum=0,decoder.ipraw.invalid_ip_version=0,tcp.reassembly_memuse=12332832,app_layer.flow.dns_tcp=0,decoder.udp=0,dns.memuse=0,flow.emerg_mode_over=0,flow_mgr.est_pruned=0,flow_mgr.rows_busy=0,flow_mgr.bypassed_pruned=0,decoder.ltnull.pkt_too_small=0,tcp.no_flow=0,defrag.max_frag_hits=0,app_layer.tx.smtp=0,decoder.invalid=0,defrag.ipv4.timeouts=0,flow.memcap=0 1501764060000000000
```
