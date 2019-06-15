# Ping Native Input Plugin

Sends a ping message and reports the results. This is done in pure go, eliminating the need to execute the system `ping` command.

There is currently no support for TTL on windows, track progress https://github.com/golang/go/issues/7175 and https://github.com/golang/go/issues/7174


### Configuration:

```toml
[[inputs.ping_native]]
  ## List of hosts to ping.
  hosts = ["8.8.8.8"]

  ## Number of pings to send per collection.
  # count = 1

  ## Interval, in s, at which to ping.
  # ping_interval = 1.0

  ## Per-ping timeout, in s (0 == no timeout).
  # timeout = 1.0

  ## Total-ping deadline, in s. Set to value equal to or lower than agent interval.
  # deadline = 10

  ## Interface or source address to send ping from.
  # interface = ""

	## Whether to ping ipv6 addresses.
  # ipv6 = false
```

#### Permission Caveat (non Windows)

Since this plugin listens on unprivileged raw sockets on linux and darwin, the system group of the user running telegraf must be allowed to create ICMP Echo sockets. [See man pages icmp(7) for `ping_group_range`](http://man7.org/linux/man-pages/man7/icmp.7.html) for linux and [man pages icmp(4) for ``](https://www.freebsd.org/cgi/man.cgi?query=icmp&apropos=0&sektion=0&manpath=Darwin+8.0.1%2Fppc&format=html) for darwin.

```
sudo sysctl -w net.ipv4.ping_group_range="GROUPID   GROUPID"
```

### Metrics:

- ping
  - tags:
    - ip
  - fields:
    - packets_transmitted (integer)
    - packets_received (integer)
    - percent_packets_loss (float)
    - ttl (integer, Not available on Windows)
    - average_response_ms (integer)
    - minimum_response_ms (integer)
    - maximum_response_ms (integer)
    - standard_deviation_ms (integer)
    - result_code (int, success = 0, no such host = 1, ping error = 2)

##### reply_received vs packets_received
<!-- todo: verify this -->

On Windows systems, "Destination net unreachable" reply will increment `packets_received` but not `reply_received`.

### Example Output:

**Windows:**
```
ping,ip=127.0.0.1 average_response_ms=94.738085,maximum_response_ms=94.790376,minimum_response_ms=94.702181,packets_received=3i,packets_transmitted=3i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.037823 1560553382000000000
```

**Linux:**
```
ping,ip=2600:: average_response_ms=94.738085,maximum_response_ms=94.790376,minimum_response_ms=94.702181,packets_received=3i,packets_transmitted=3i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.037823,ttl=52i 1560553382000000000
```


```toml
[agent]
  interval="5s"
  flush_interval="1s"
  omit_hostname=true

[[outputs.file]]

[[inputs.ping_native]]
  count = 3
  ping_interval = 1.0
  timeout = 1.0
  deadline = 10
  interface= "2604:a880:1:20::352:7001"
  ipv6=true
  hosts = [
    "google.com",
    "2600::",
    "2620:0:ccc::2",
    "2620:0:ccd::2",
  ]
```

<!-- ```
root@ipv6-test:~# ./telegraf --config ./tel.conf --test
2019-06-14T23:02:59Z I! Starting Telegraf 
> ping,ip=2620:0:ccc::2 average_response_ms=1.857576,maximum_response_ms=2.098475,minimum_response_ms=1.724104,packets_received=3i,packets_transmitted=3i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.170672,ttl=61i 1560553382000000000
> ping,ip=2620:0:ccd::2 average_response_ms=1.769044,maximum_response_ms=2.004256,minimum_response_ms=1.639915,packets_received=3i,packets_transmitted=3i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.166585,ttl=61i 1560553382000000000
> ping,ip=2607:f8b0:4005:80b::200e average_response_ms=3.000772,maximum_response_ms=3.14064,minimum_response_ms=2.905394,packets_received=3i,packets_transmitted=3i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.10106,ttl=57i 1560553382000000000
> ping,ip=2600:: average_response_ms=94.738085,maximum_response_ms=94.790376,minimum_response_ms=94.702181,packets_received=3i,packets_transmitted=3i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.037823,ttl=52i 1560553382000000000
``` -->