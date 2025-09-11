# Syslog Input Plugin

This service plugin listens for [syslog][syslog] messages transmitted over a
Unix Domain socket, [UDP][rfc5426], [TCP][rfc6587] or [TLS][rfc5425] with or
without the octet counting framing.

Syslog messages should be formatted according to the [syslog protocol][rfc5424]
or the [BSD syslog protocol][rfc3164].

⭐ Telegraf v1.7.0
🏷️ logging
💻 all

[syslog]: https://en.wikipedia.org/wiki/Syslog
[rfc5426]: https://tools.ietf.org/html/rfc5426
[rfc6587]: https://tools.ietf.org/html/rfc6587
[rfc5425]: https://tools.ietf.org/html/rfc5425
[rfc5424]: https://tools.ietf.org/html/rfc5424
[rfc3164]: https://tools.ietf.org/html/rfc3164

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listen and wait for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
[[inputs.syslog]]
  ## Protocol, address and port to host the syslog receiver.
  ## If no host is specified, then localhost is used.
  ## If no port is specified, 6514 is used (RFC5425#section-4.1).
  ##   ex: server = "tcp://localhost:6514"
  ##       server = "udp://:6514"
  ##       server = "unix:///var/run/telegraf-syslog.sock"
  ## When using tcp, consider using 'tcp4' or 'tcp6' to force the usage of IPv4
  ## or IPV6 respectively. There are cases, where when not specified, a system
  ## may force an IPv4 mapped IPv6 address.
  server = "tcp://127.0.0.1:6514"

  ## Permission for unix sockets (only available on unix sockets)
  ## This setting may not be respected by some platforms. To safely restrict
  ## permissions it is recommended to place the socket into a previously
  ## created directory with the desired permissions.
  ##   ex: socket_mode = "777"
  # socket_mode = ""

  ## Maximum number of concurrent connections (only available on stream sockets like TCP)
  ## Zero means unlimited.
  # max_connections = 0

  ## Read timeout (only available on stream sockets like TCP)
  ## Zero means unlimited.
  # read_timeout = "0s"

  ## Optional TLS configuration (only available on stream sockets like TCP)
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key  = "/etc/telegraf/key.pem"
  ## Enables client authentication if set.
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Maximum socket buffer size (in bytes when no unit specified)
  ## For stream sockets, once the buffer fills up, the sender will start
  ## backing up. For datagram sockets, once the buffer fills up, metrics will
  ## start dropping. Defaults to the OS default.
  # read_buffer_size = "64KiB"

  ## Period between keep alive probes (only applies to TCP sockets)
  ## Zero disables keep alive probes. Defaults to the OS configuration.
  # keep_alive_period = "5m"

  ## Content encoding for message payloads
  ## Can be set to "gzip" for compressed payloads or "identity" for no encoding.
  # content_encoding = "identity"

  ## Maximum size of decoded packet (in bytes when no unit specified)
  # max_decompression_size = "500MB"

  ## Framing technique used for messages transport
  ## Available settings are:
  ##   octet-counting  -- see RFC5425#section-4.3.1 and RFC6587#section-3.4.1
  ##   non-transparent -- see RFC6587#section-3.4.2
  # framing = "octet-counting"

  ## The trailer to be expected in case of non-transparent framing (default = "LF").
  ## Must be one of "LF", or "NUL".
  # trailer = "LF"

  ## Whether to parse in best effort mode or not (default = false).
  ## By default best effort parsing is off.
  # best_effort = false

  ## The RFC standard to use for message parsing
  ## By default RFC5424 is used. RFC3164 only supports UDP transport (no streaming support)
  ## Must be one of "RFC5424", or "RFC3164".
  # syslog_standard = "RFC5424"

  ## Character to prepend to SD-PARAMs (default = "_").
  ## A syslog message can contain multiple parameters and multiple identifiers within structured data section.
  ## Eg., [id1 name1="val1" name2="val2"][id2 name1="val1" nameA="valA"]
  ## For each combination a field is created.
  ## Its name is created concatenating identifier, sdparam_separator, and parameter name.
  # sdparam_separator = "_"

  ## Maximum length allowed for a single message (in bytes when no unit specified)
  ## Only applies to octet-counting framing.
  # max_message_length = "8KiB"
```

### Message transport

The `framing` option only applies to streams. It governs the way we expect to
receive messages within the stream.  Namely, with the [`"octet counting"`][1]
technique (default) or with the [`"non-transparent"`][2] framing.

The `trailer` option only applies when `framing` option is
`"non-transparent"`. It must have one of the following values: `"LF"` (default),
or `"NUL"`.

[1]: https://tools.ietf.org/html/rfc5425#section-4.3

[2]: https://tools.ietf.org/html/rfc6587#section-3.4.2

### Best effort

The [`best_effort`](https://github.com/influxdata/go-syslog#best-effort-mode)
option instructs the parser to extract partial but valid info from syslog
messages. If unset only full messages will be collected.

### Rsyslog Integration

Rsyslog can be configured to forward logging messages to Telegraf by configuring
[remote logging][remote_logging].

Most system are setup with a configuration split between `/etc/rsyslog.conf`
and the files in the `/etc/rsyslog.d/` directory, it is recommended to add the
new configuration into the config directory to simplify updates to the main
config file.

Add the following lines to `/etc/rsyslog.d/50-telegraf.conf` making
adjustments to the target address as needed:

```shell
$ActionQueueType LinkedList # use asynchronous processing
$ActionQueueFileName srvrfwd # set file name, also enables disk mode
$ActionResumeRetryCount -1 # infinite retries on insert failure
$ActionQueueSaveOnShutdown on # save in-memory data if rsyslog shuts down

# forward over tcp with octet framing according to RFC 5425
*.* @@(o)127.0.0.1:6514;RSYSLOG_SyslogProtocol23Format

# uncomment to use udp according to RFC 5424
#*.* @127.0.0.1:6514;RSYSLOG_SyslogProtocol23Format
```

You can alternately use `advanced` format (aka RainerScript):

```bash
# forward over tcp with octet framing according to RFC 5425
action(type="omfwd" Protocol="tcp" TCP_Framing="octet-counted" Target="127.0.0.1" Port="6514" Template="RSYSLOG_SyslogProtocol23Format")

# uncomment to use udp according to RFC 5424
#action(type="omfwd" Protocol="udp" Target="127.0.0.1" Port="6514" Template="RSYSLOG_SyslogProtocol23Format")
```

To complete TLS setup please refer to [rsyslog docs][rsyslog_docs].

[remote_logging]: https://www.rsyslog.com/doc/v8-stable/configuration/actions.html#remote-machine
[rsyslog_docs]: https://www.rsyslog.com/doc/v8-stable/tutorials/tls.html

## Troubleshooting

```sh
# TCP with octet framing
echo "57 <13>1 2018-10-01T12:00:00.0Z example.org root - - - test" | nc 127.0.0.1 6514

# UDP
echo "<13>1 2018-10-01T12:00:00.0Z example.org root - - - test" | nc -u 127.0.0.1 6514
```

### Resolving Source IPs

The `source` tag stores the remote IP address of the syslog sender.
To resolve these IPs to DNS names, use the
[`reverse_dns` processor][plugin_reverse_dns]

You can send debugging messages directly to the input plugin using netcat:

[plugin_reverse_dns]: /plugins/processors/reverse_dns/README.md

### RFC3164

RFC3164 encoded messages are supported for UDP only, but not all vendors output
valid RFC3164 messages by default

- E.g. Cisco IOS

If you see the following error, it is due to a message encoded in this format:

 ```shell
 E! Error in plugin [inputs.syslog]: expecting a version value in the range 1-999 [col 5]
 ```

Users can use rsyslog to translate RFC3164 syslog messages into RFC5424 format.
Add the following lines to the rsyslog configuration file
(e.g. `/etc/rsyslog.d/50-telegraf.conf`):

```s
# This makes rsyslog listen on 127.0.0.1:514 to receive RFC3164 udp
# messages which can them be forwarded to telegraf as RFC5424
$ModLoad imudp #loads the udp module
$UDPServerAddress 127.0.0.1
$UDPServerRun 514
```

Make adjustments to the target address as needed and sent your RFC3164 messages
to port 514.

## Metrics

- syslog
  - tags
    - severity (string)
    - facility (string)
    - hostname (string)
    - appname (string)
    - source (string)
  - fields
    - version (integer)
    - severity_code (integer)
    - facility_code (integer)
    - timestamp (integer): the time recorded in the syslog message
    - procid (string)
    - msgid (string)
    - sdid (bool)
    - *Structured Data* (string)
  - timestamp: the time the messages was received

*Structured data* produces field keys by combining the `SD_ID` with the
`PARAM_NAME` combined using the `sdparam_separator` as in the following example:

```shell
170 <165>1 2018-10-01:14:15.000Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"] An application event log entry...
```

```shell
syslog,appname=evntslog,facility=local4,hostname=mymachine.example.com,severity=notice exampleSDID@32473_eventID="1011",exampleSDID@32473_eventSource="Application",exampleSDID@32473_iut="3",facility_code=20i,message="An application event log entry...",msgid="ID47",severity_code=5i,timestamp=1065910455003000000i,version=1i 1538421339749472344
```

## Example Output

Here is example output of this plugin:

```text
syslog,appname=docker-compose,facility=daemon,host=bb8,hostname=droplet,location=home,severity=info,source=10.0.0.12 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643706396113000i,version=1i 1624643706400667198
syslog,appname=tailscaled,facility=daemon,host=bb8,hostname=dev,location=home,severity=info,source=10.0.0.15 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643706403394000i,version=1i 1624643706407850408
syslog,appname=docker-compose,facility=daemon,host=bb8,hostname=droplet,location=home,severity=info,source=10.0.0.12 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643706675853000i,version=1i 1624643706679251683
syslog,appname=telegraf,facility=daemon,host=bb8,hostname=droplet,location=home,severity=info,source=10.0.0.12 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643710005006000i,version=1i 1624643710008285426
syslog,appname=telegraf,facility=daemon,host=bb8,hostname=droplet,location=home,severity=info,source=10.0.0.12 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643710005696000i,version=1i 1624643710010754050
syslog,appname=docker-compose,facility=daemon,host=bb8,hostname=droplet,location=home,severity=info,source=10.0.0.12 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643715777813000i,version=1i 1624643715782158154
syslog,appname=docker-compose,facility=daemon,host=bb8,hostname=droplet,location=home,severity=info,source=10.0.0.12 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643716396547000i,version=1i 1624643716400395788
syslog,appname=tailscaled,facility=daemon,host=bb8,hostname=dev,location=home,severity=info,source=10.0.0.15 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643716404931000i,version=1i 1624643716416947058
syslog,appname=docker-compose,facility=daemon,host=bb8,hostname=droplet,location=home,severity=info,source=10.0.0.12 facility_code=3i,message="<redacted>",severity_code=6i,timestamp=1624643716676633000i,version=1i 1624643716680157558
```
