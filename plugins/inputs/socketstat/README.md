# SocketStat plugin

The socketstat plugin gathers indicators from established connections, using iproute2's `ss` command.

The `ss` command does not require specific privileges.

**WARNING: The output format will produce series with very high cardinality.** You should either store those by an engine which doesn't suffer from it, use a short retention policy or do appropriate filtering.

## Configuration

```toml
# Gather indicators from established connections, using iproute2's ss command.
[[inputs.socketstat]]
  ## ss can display information about tcp, udp, raw, unix, packet, dccp and sctp sockets
  ## Specify here the types you want to gather
  socket_types = [ "tcp", "udp" ]
  ## The default timeout of 1s for ss execution can be overridden here:
  # timeout = "1s"
```

## Measurements & Fields

- socketstat
  - state (string) (for tcp, dccp and sctp protocols)
  - If ss provides it (it depends on the protocol and ss version):
    - bytes_acked (integer, bytes)
    - bytes_received (integer, bytes)
    - segs_out (integer, count)
    - segs_in (integer, count)
    - data_segs_out (integer, count)
    - data_segs_in (integer, count)

## Tags

- All measurements have the following tags:
  - proto
  - local_addr
  - local_port
  - remote_addr
  - remote_port

## Example Output

### recent ss version (iproute2 4.3.0 here)

```sh
./telegraf --config telegraf.conf --input-filter socketstat --test
> socketstat,host=ubuntu-xenial,local_addr=10.6.231.226,local_port=42716,proto=tcp,remote_addr=192.168.2.21,remote_port=80 bytes_acked=184i,bytes_received=2624519595i,recv_q=4344i,segs_in=1812580i,segs_out=661642i,send_q=0i,state="ESTAB" 1606457205000000000
```

### older ss version (iproute2 3.12.0 here)

```sh
./telegraf --config telegraf.conf --input-filter socketstat --test
> socketstat,host=ubuntu-trusty,local_addr=10.6.231.163,local_port=35890,proto=tcp,remote_addr=192.168.2.21,remote_port=80 recv_q=0i,send_q=0i,state="ESTAB" 1606456977000000000
```
