# syslog input plugin

Collects syslog messages as per RFC5425 or RFC5426.

It can act as a syslog transport receiver over TLS (or TCP) - ie., RFC5425 - or over UDP - ie., RFC5426.

This plugin listens for syslog messages following RFC5424 format. When received it parses them extracting metrics.

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

  ## Read timeout (default = 500ms).
  ## 0 means unlimited.
  # read_timeout = 500ms

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

#### Other configs

Other available configurations are:

- `keep_alive_period`, `max_connections` for stream sockets
- `read_timeout`
- `best_effort` to tell the parser to work until it is able to do and extract partial but valid info (more [here](https://github.com/influxdata/go-syslog#best-effort-mode))
- `sdparam_separator` to choose how to separate structured data param name from its structured data identifier

### Metrics

- syslog
  - fields
    - **version** (`uint16`)
    - **severity_code** (`int`)
    - **facility_code** (`int`)
    - timestamp (`time.Time`)
    - procid (`string`)
    - msgid (`string`)
    - *sdid* (`bool`)
    - *sdid . sdparam_separator . sdparam_name* (`string`)
  - tags
    - **severity** (`string`)
    - **facility** (`string`)
    - hostname (`string`)
    - appname (`string`)

The name of fields in _italic_ corresponds to their runtime value.

The fields/tags which name is in **bold** will always be present when a valid Syslog message has been received.

### RSYSLOG integration

The following instructions illustrate how to configure a syslog transport sender as per RFC5425 - ie., using the octect framing technique - via RSYSLOG.

Install `rsyslog`.

Give it a configuration - ie., `/etc/rsyslog.conf`.

```
$ModLoad imuxsock  # provides support for local system logging
$ModLoad imklog    # provides kernel logging support
$ModLoad immark    # provides heart-beat logs
$FileOwner root
$FileGroup root
$FileCreateMode 0640
$DirCreateMode 0755
$Umask 0022
$WorkDirectory /var/spool/rsyslog # default location for work (spool) files
$ActionQueueType LinkedList # use asynchronous processing
$ActionQueueFileName srvrfwd # set file name, also enables disk mode
$ActionResumeRetryCount -1 # infinite retries on insert failure
$ActionQueueSaveOnShutdown on # save in-memory data if rsyslog shuts down
$IncludeConfig /etc/rsyslog.d/*.conf
```

Specify you want the octet framing technique enabled and the format of each syslog message to follow the RFC5424.

Create a file - eg., `/etc/rsyslog.d/50-default.conf` - containing:

```
*.* @@(o)127.0.0.1:6514;RSYSLOG_SyslogProtocol23Format
```

To complete the TLS setup please refer to [rsyslog docs](https://www.rsyslog.com/doc/v8-stable/tutorials/tls.html).

Notice that this configuration tells `rsyslog` to broadcast messages to `127.0.0.1>6514`.

So you have to configure this plugin accordingly.