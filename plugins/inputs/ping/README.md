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

  ## Specify the ping executable binary, default is "ping"
  # binary = "ping"

  ## Arguments for ping command
  ## when arguments is not empty, other options (ping_interval, timeout, etc) will be ignored
  # arguments = ["-c", "3"]
```

#### File Limit

Since this plugin runs the ping command, it may need to open several files per
host.  With a large host list you may receive a `too many open files` error.

To increase this limit on platforms using systemd it must be done in the
service file.


Find the service unit file:
```
$ systemctl show telegraf.service -p FragmentPath
FragmentPath=/lib/systemd/system/telegraf.service
```

Set the file number limit:
```
[Service]
LimitNOFILE=4096
```

### Metrics:

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
