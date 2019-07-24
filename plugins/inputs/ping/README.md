# Ping Input Plugin

Sends a ping message by executing the system ping command and reports the results.

Most ping command implementations are supported, one notable exception being
that there is currently no support for GNU Inetutils ping.  You may instead
use the iputils-ping implementation:
```
apt-get install iputils-ping
```

When using `method = "native"` a ping is sent and the results are reported in pure go, eliminating the need to execute the system `ping` command. Not using the system binary allows the use of this plugin on non-english systems.

There is currently no support for TTL on windows with `"native"`; track progress at https://github.com/golang/go/issues/7175 and https://github.com/golang/go/issues/7174

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

  ## How to ping. "native" doesn't have external dependencies, while "exec" depends on 'ping'.
  # method = "exec"

  ## Specify the ping executable binary, default is "ping"
  # binary = "ping"

  ## Arguments for ping command. When arguments is not empty, system binary will be used and
  ## other options (ping_interval, timeout, etc) will be ignored
  # arguments = ["-c", "3"]

  ## Use only ipv6 addresses when resolving hostnames.
  # ipv6 = false
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

#### Permission Caveat (non Windows)

It is preferred that this plugin listen on privileged ICMP sockets. To do so, telegraf can either be run as the root user or the root user can add the capability to access raw sockets to telegraf by running the following commant:

```
setcap cap_net_raw=eip /path/to/telegraf
```

Another option (doesn't work as well or in all circumstances) is to listen on unprivileged raw sockets (non-Windows only). The system group of the user running telegraf must be allowed to create ICMP Echo sockets. [See man pages icmp(7) for `ping_group_range`](http://man7.org/linux/man-pages/man7/icmp.7.html). On Linux hosts, run the following to give a group the proper permissions:

```
sudo sysctl -w net.ipv4.ping_group_range="GROUP_ID_LOW   GROUP_ID_HIGH"
```


### Metrics:

- ping
  - tags:
    - url
  - fields:
    - packets_transmitted (integer)
    - packets_received (integer)
    - percent_packets_loss (float)
    - ttl (integer, Not available on Windows)
    - average_response_ms (integer)
    - minimum_response_ms (integer)
    - maximum_response_ms (integer)
    - standard_deviation_ms (integer, Available on Windows only with native ping)
    - errors (float, Windows only)
    - reply_received (integer, Windows only*)
    - percent_reply_loss (float, Windows only*)
    - result_code (int, success = 0, no such host = 1, ping error = 2)

##### reply_received vs packets_received

On Windows systems, "Destination net unreachable" reply will increment `packets_received` but not `reply_received`*

### Example Output:

**Windows:**
```
ping,url=example.org result_code=0i,average_response_ms=7i,maximum_response_ms=9i,minimum_response_ms=7i,packets_received=4i,packets_transmitted=4i,percent_packet_loss=0,percent_reply_loss=0,reply_received=4i 1469879119000000000
```

**Linux:**
```
ping,url=example.org average_response_ms=23.066,ttl=63,maximum_response_ms=24.64,minimum_response_ms=22.451,packets_received=5i,packets_transmitted=5i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.809 1535747258000000000
```

*not when `method = "native"` is used
