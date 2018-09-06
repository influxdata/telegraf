# Ping Input Plugin

Sends a ping message by executing the system ping command and reports the results.

Currently there is no support for GNU Inetutils, use with iputils-ping
instead:
```
apt-get install iputils-ping
```

### Configuration:

```toml
[[inputs.ping]]
  ## List of urls to ping
  urls = ["example.org"]

  ## Number of pings to send per collection (ping -c <COUNT>)
  # count = 1

  ## Interval, in s, at which to ping. 0 == default (ping -i <PING_INTERVAL>)
  ## Not available in Windows.
  # ping_interval = 1.0

  ## Per-ping timeout, in s. 0 == no timeout (ping -W <TIMEOUT>)
  # timeout = 1.0

  ## Total-ping deadline, in s. 0 == no deadline (ping -w <DEADLINE>)
  # deadline = 10

  ## Interface or source address to send ping from (ping -I <INTERFACE/SRC_ADDR>)
  ## on Darwin and Freebsd only source address possible: (ping -S <SRC_ADDR>)
  # interface = ""
```

### Metrics:

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
    - 2: ping error
- ping
  - tags:
    - url
  - fields:
    - packets_transmitted (integer)
    - packets_received (integer)
    - percent_packets_loss (float)
    - average_response_ms (integer)
    - minimum_response_ms (integer)
    - maximum_response_ms (integer)
    - standard_deviation_ms (integer, Not available on Windows)
    - errors (float, Windows only)
    - reply_received (integer, Windows only)
    - percent_reply_loss (float, Windows only)
    - result_code (int, success = 0, no such host = 1, ping error = 2)

##### reply_received vs packets_received

On Windows systems, "Destination net unreachable" reply will increment `packets_received` but not `reply_received`.

### Example Output:

**Windows:**
```
ping,url=example.org result_code=0i,average_response_ms=7i,maximum_response_ms=9i,minimum_response_ms=7i,packets_received=4i,packets_transmitted=4i,percent_packet_loss=0,percent_reply_loss=0,reply_received=4i 1469879119000000000
```

**Linux:**
```
ping,url=example.org average_response_ms=23.066,maximum_response_ms=24.64,minimum_response_ms=22.451,packets_received=5i,packets_transmitted=5i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.809 1535747258000000000
```
