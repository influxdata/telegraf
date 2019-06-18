# Ping Native Input Plugin

Sends a ping message and reports the results. This is done in pure go, eliminating the need to execute the system `ping` command.

There is currently no support for TTL on windows, track progress at https://github.com/golang/go/issues/7175 and https://github.com/golang/go/issues/7174


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

Since this plugin listens on unprivileged raw sockets on Linux, the system group of the user running telegraf must be allowed to create ICMP Echo sockets. [See man pages icmp(7) for `ping_group_range`](http://man7.org/linux/man-pages/man7/icmp.7.html). On Linux hosts, run the following to give a group the proper permissions:

```
sudo sysctl -w net.ipv4.ping_group_range="GROUPID   GROUPID"
```

### Metrics:

- ping
  - tags:
    - source
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

### Example Output:

**Windows:**
```
ping,source=google.com average_response_ms=94.738085,maximum_response_ms=94.790376,minimum_response_ms=94.702181,packets_received=3i,packets_transmitted=3i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.037823 1560553382000000000
```

**Other:**
```
ping,source=google.com average_response_ms=94.738085,maximum_response_ms=94.790376,minimum_response_ms=94.702181,packets_received=3i,packets_transmitted=3i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.037823,ttl=52i 1560553382000000000
```
