# SocketStat plugin

The socketstat plugin gathers indicators from established connections, using iproute2's ss command.

The ss command does not require specific privileges.

### Configuration

```toml
  ## ss can display information about tcp, udp, raw, unix, packet, dccp and sctp sockets
  ## Specify here the types you want to gather
  socket_types = [ "tcp", "udp", "raw" ]
```

### Measurements & Fields:

- socketstat
    - state (string)
    - proto (string)
    - For tcp, if ss provides it (depends on the version):
        - bytes_acked (integer, bytes)
        - bytes_received (integer, bytes)
        - segs_out (integer, count)
        - segs_in (integer, count)
        - data_segs_out (integer, count)
        - data_segs_in (integer, count)

### Tags:

- All measurements have the following tags:
    - local_addr
    - local_port
    - remote_addr
    - remote_port
    - state
 
### Example Output:

#### recent ss version

```
$ ss -ein
Netid  State      Recv-Q Send-Q Local Address:Port               Peer Address:Port              
tcp   ESTAB      0      0          192.168.1.21:53896                  1.2.3.4:443                 timer:(keepalive,40sec,0) uid:1000 ino:88276 sk:1c1 <->
	 ts sack cubic wscale:9,7 rto:208 rtt:6.367/1.849 ato:40 mss:1448 cwnd:10 bytes_acked:2155 bytes_received:3873216 segs_out:218 segs_in:2742 data_segs_out:17 data_segs_in:2730 send 18.2Mbps lastsnd:411100 lastrcv:411084 lastack:4252 pacing_rate 36.4Mbps rcv_rtt:8.451 rcv_space:359117 minrtt:4.004
```

```$ ./telegraf --config telegraf.conf --input-filter socketstat --test
socketstat,proto=tcp,local_addr=192.168.1.21,local_port=53896i,remote_addr=1.2.3.4,remote_port=443i state=ESTAB,bytes_acked=2155i,bytes_received=3873216i,segs_out=218i,segs_in=2742i,data_segs_out=17i,data_segs_in=2730i
```

#### older ss version

tcp   ESTAB      0      0                                                  192.168.1.21:38776                                           1.2.3.4:18080  uid:100 ino:378042749 sk:ffff880257013100 <->
	 ts sack cubic wscale:7,7 rto:284 rtt:84.5/8 ato:40 mss:1448 cwnd:5 ssthresh:3 send 685.4Kbps retrans:0/4 rcv_rtt:88 rcv_space:29200
socketstat,proto=tcp,local_addr=192.168.1.21,local_port=38776i,remote_addr=1.2.3.4,remote_port=18080i,state=ESTAB
