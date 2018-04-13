# Syslog RFC5424 Parsing

## Configure RSYSLOG
In `/etc/rsyslog.d/50-default.conf` add:


```
*.*         @127.0.0.1:5140;RSYSLOG_SyslogProtocol23Format
```

## Telegraf config

```toml
[[inputs.socket_listener]]
  service_address = "udp://:5140"
  data_format = "syslog"
```
