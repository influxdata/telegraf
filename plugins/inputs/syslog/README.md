# syslog input plugin

Collects syslog messages as per RFC5425 or RFC5426.

It can act as a syslog transport receiver over TLS (or TCP) - ie., RFC5425 - or over UDP - ie., RFC5426.

This plugin listens for syslog messages following RFC5424 format. When received it parses them extracting metrics.

### Configuration:

#### TCP

The minimal configuration is the following:

```toml
[[inputs.syslog]]
  address = ":6514"
```

This starts this plugins as a syslog receiver over TCP protocol on port 6514.

#### TLS

To configure it as a TLS syslog receiver as recommended by RFC5425 give it the following configuration:

```toml
[[inputs.syslog]]
  address = ":6514"
  tls_cacert = "/etc/telegraf/ca.pem"
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"
```

#### UDP

To configure this plugin as per RFC5426 give it the following configuration:

```toml
[[inputs.syslog]]
  protocol = "udp"
  address = ":6514"
```

#### Other configs

Other available configurations are:

- `keep_alive_period`, `max_connections` for stream sockets
- `best_effort` to tell the parser to work until it is able to do and extract partial but valid info

### Measurements & Fields

- syslog
  - **version** (`uint16`)
  - timestamp (`time.Time`)
  - procid (`string`)
  - msgid (`string`)
  - _structureddata element id_ (`bool`)
  - _structureddata element parameter name_ (`string`)

The name of fields in _italic_ corresponds to their runtime value.

The fields which name is in **bold** will always be present when a valid Syslog message has been received.

### Tags

- **severity** (`string`)
- **severity_level** (`string`)
- **facility** (`string`)
- **facility_message** (`string`)
- hostname (`string`)
- appname (`string`)

The tags which name is in **bold** will always be present when a valid Syslog message has been received.

### Syslog transport sender

The following instructions illustrate how to configure a syslog transport sender as per RFC5425 - ie., using the octect framing technique.

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