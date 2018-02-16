# Ping Input plugin

This input plugin will measures the round-trip

### Configuration:

```
# NOTE: this plugin forks the ping command. You may need to set capabilities
# via setcap cap_net_raw+p /bin/ping
[[inputs.ping]]
## List of urls to ping
urls = ["www.google.com"] # required
## number of pings to send per collection (ping -c <COUNT>)
# count = 1
## interval, in s, at which to ping. 0 == default (ping -i <PING_INTERVAL>)
## Not available in Windows.
# ping_interval = 1.0
## per-ping timeout, in s. 0 == no timeout (ping -W <TIMEOUT>)
# timeout = 1.0
## total-ping deadline, in s. 0 == no deadline (ping -w <DEADLINE>)
# deadline = 10
## interface or source address to send ping from (ping -I <INTERFACE/SRC_ADDR>)
## on Darwin and Freebsd only source address possible: (ping -S <SRC_ADDR>)
# interface = ""
```

### Measurements & Fields:

- packets_transmitted ( from ping output )
- reply_received ( increasing only on valid metric from echo replay, eg. 'Destination net unreachable' reply will increment packets_received but not reply_received )
- packets_received ( from ping output )
- percent_reply_loss ( compute from packets_transmitted and reply_received )
- percent_packets_loss ( compute from packets_transmitted and packets_received )
- errors ( when host can not be found or wrong parameters is passed to application )
- response time
    - average_response_ms ( compute from minimum_response_ms and maximum_response_ms )
    - minimum_response_ms ( from ping output )
    - maximum_response_ms ( from ping output )
- result_code
    - 0: success
    - 1: no such host

### Tags:

- host
- url

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter ping --test
* Plugin: ping, Collection 1
ping,host=WIN-PBAPLP511R7,url=www.google.com result_code=0i,average_response_ms=7i,maximum_response_ms=9i,minimum_response_ms=7i,packets_received=4i,packets_transmitted=4i,percent_packet_loss=0,percent_reply_loss=0,reply_received=4i 1469879119000000000
```
