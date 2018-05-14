# syslog input plugin

Collect syslog messages as per RFC5425.

It acts as a syslog transport receiver.

This plugin listens for syslog messages. When received it parses them extracting metrics.

### Configuration:

```toml
[[inputs.syslog]]
  address = ":6514"
```

_TODO: TLS config_.

### Measurements & Fields:

_TODO_

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

File `/etc/rsyslog.d/50-default.conf`.

```
*.* @@(o)127.0.0.1:6514;RSYSLOG_SyslogProtocol23Format
```

_TODO: TLS config_.

Notice that this configuration tells `rsyslog` to broadcast messages to `127.0.0.1>6514`.

So you have to configure this plugin accordingly.