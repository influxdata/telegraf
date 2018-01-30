# SocketStat plugin

The socketstat plugin gathers indicators from established connections, using iproute2's `ss` command.

The ss command does not require specific privileges.

The output format will have a very high cardinality. It should either be stored by an engine which doesn't suffer from it, of with a short retention policy.

### Configuration

```toml
[[inputs.socketstat]]
  ## ss can display information about tcp, udp, raw, unix, packet, dccp and sctp sockets
  ## Specify here the types you want to gather
  socket_types = [ "tcp", "udp" ]
```

### Measurements & Fields:

- socketstat
    - state (string) (tcp, dccp and sctp)
    - If ss provides it (it depends on the protocol and ss version):
        - bytes_acked (integer, bytes)
        - bytes_received (integer, bytes)
        - segs_out (integer, count)
        - segs_in (integer, count)
        - data_segs_out (integer, count)
        - data_segs_in (integer, count)

### Tags:

- All measurements have the following tags:
    - proto
    - state (tcp, dccp and sctp)
    - local_addr
    - local_port
    - remote_addr
    - remote_port
 
### Example Output:

#### recent ss version

```
$ ss -in --tcp
State      Recv-Q Send-Q Local Address:Port               Peer Address:Port
ESTAB      0      0      127.0.0.1:53692              127.0.0.1:7778
	 cubic wscale:7,7 rto:224 rtt:21.55/14.181 ato:40 mss:25600 cwnd:10 bytes_acked:34525 bytes_received:2663883 segs_out:2331 segs_in:2136 data_segs_out:654 data_segs_in:1680 send 95.0Mbps lastsnd:15112 lastrcv:15084 lastack:15084 pacing_rate 190.1Mbps rcv_rtt:93.021 rcv_space:60281 minrtt:0.028

./telegraf --config telegraf.conf --input-filter socketstat --test
> socketstat,proto=tcp,local_addr=127.0.0.1,local_port=53692,remote_addr=127.0.0.1,remote_port=7778,host=mymachine bytes_acked=34525i,segs_out=2331i,data_segs_out=654i,data_segs_in=1680i,send_q=0i,recv_q=0i,bytes_received=2663883i,segs_in=2136i,state="ESTAB" 1515496754000000000
```

#### older ss version

```
$ ss -in --tcp
State       Recv-Q Send-Q                           Local Address:Port                             Peer Address:Port
ESTAB      0      0                                                  192.168.1.21:38776                                           1.2.3.4:18080
	 cubic wscale:7,7 rto:284 rtt:84.5/8 ato:40 cwnd:5 ssthresh:3 send 685.4Kbps rcv_rtt:88 rcv_space:29200

./telegraf --config telegraf.conf --input-filter socketstat --test
socketstat,proto=tcp,local_addr=1.2.3.4,local_port=18080,remote_addr=192.168.1.21,remote_port=38776,state=ESTAB
```
