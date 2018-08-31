# Syslog Input Plugin

The syslog plugin listens for syslog messages transmitted over
[UDP](https://tools.ietf.org/html/rfc5426) or
[TCP](https://tools.ietf.org/html/rfc5425).

Syslog messages should be formatted according to
[RFC 5424](https://tools.ietf.org/html/rfc5424).

### Configuration

```toml
[[inputs.syslog]]
  ## Specify an ip or hostname with port - eg., tcp://localhost:6514, tcp://10.0.0.1:6514
  ## Protocol, address and port to host the syslog receiver.
  ## If no host is specified, then localhost is used.
  ## If no port is specified, 6514 is used (RFC5425#section-4.1).
  server = "tcp://:6514"

  ## TLS Config
  # tls_allowed_cacerts = ["/etc/telegraf/ca.pem"]
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Period between keep alive probes.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  ## Only applies to stream sockets (e.g. TCP).
  # keep_alive_period = "5m"

  ## Maximum number of concurrent connections (default = 0).
  ## 0 means unlimited.
  ## Only applies to stream sockets (e.g. TCP).
  # max_connections = 1024

  ## Read timeout is the maximum time allowed for reading a single message (default = 5s).
  ## 0 means unlimited.
  # read_timeout = "5s"

  ## Whether to parse in best effort mode or not (default = false).
  ## By default best effort parsing is off.
  # best_effort = false

  ## Character to prepend to SD-PARAMs (default = "_").
  ## A syslog message can contain multiple parameters and multiple identifiers within structured data section.
  ## Eg., [id1 name1="val1" name2="val2"][id2 name1="val1" nameA="valA"]
  ## For each combination a field is created.
  ## Its name is created concatenating identifier, sdparam_separator, and parameter name.
  # sdparam_separator = "_"
```

#### Best Effort

The [`best_effort`](https://github.com/influxdata/go-syslog#best-effort-mode)
option instructs the parser to extract partial but valid info from syslog
messages.  If unset only full messages will be collected.

### Metrics

- syslog
  - tags
    - severity (string)
    - facility (string)
    - hostname (string)
    - appname (string)
  - fields
    - version (integer)
    - severity_code (integer)
    - facility_code (integer)
    - timestamp (integer)
    - procid (string)
    - msgid (string)
    - sdid (bool)
    - *Structured Data* (string)

### Rsyslog Integration

Rsyslog can be configured to forward logging messages to Telegraf by configuring
[remote logging](https://www.rsyslog.com/doc/v8-stable/configuration/actions.html#remote-machine).

Most system are setup with a configuration split between `/etc/rsyslog.conf`
and the files in the `/etc/rsyslog.d/` directory, it is recommended to add the
new configuration into the config directory to simplify updates to the main
config file.

Add the following lines to `/etc/rsyslog.d/50-telegraf.conf` making
adjustments to the target address as needed:
```
$ActionQueueType LinkedList # use asynchronous processing
$ActionQueueFileName srvrfwd # set file name, also enables disk mode
$ActionResumeRetryCount -1 # infinite retries on insert failure
$ActionQueueSaveOnShutdown on # save in-memory data if rsyslog shuts down

# forward over tcp with octet framing according to RFC 5425
*.* @@(o)127.0.0.1:6514;RSYSLOG_SyslogProtocol23Format
```

To complete TLS setup please refer to [rsyslog docs](https://www.rsyslog.com/doc/v8-stable/tutorials/tls.html).
