# Telegraf plugin: Apache

#### Plugin arguments:
- **urls** []string: List of apache-status URLs to collect from. Default is "http://localhost/server-status?auto".
- **username** string: Username for HTTP basic authentication
- **password** string: Password for HTTP basic authentication
- **timeout** duration: time that the HTTP connection will remain waiting for response. Default 4 seconds ("4s")

##### Optional SSL Config

- **ssl_ca** string: the full path for the SSL CA certicate
- **ssl_cert** string: the full path for the SSL certificate
- **ssl_key** string: the full path for the key file
- **insecure_skip_verify** bool: if true HTTP client will skip all SSL verifications related to peer and host. Default to false

#### Description

The Apache plugin collects from the /server-status?auto URL. See
[apache.org/server-status?auto](http://www.apache.org/server-status?auto) for an
example. And
[here](http://httpd.apache.org/docs/2.2/mod/mod_status.html) for the apache
mod_status documentation.

# Measurements:

Meta:
- tags: `port=<port>`, `server=url`

- apache_TotalAccesses
- apache_TotalkBytes
- apache_CPULoad
- apache_Uptime
- apache_ReqPerSec
- apache_BytesPerSec
- apache_BytesPerReq
- apache_BusyWorkers
- apache_IdleWorkers
- apache_ConnsTotal
- apache_ConnsAsyncWriting
- apache_ConnsAsyncKeepAlive
- apache_ConnsAsyncClosing

### Scoreboard measurements

- apache_scboard_waiting
- apache_scboard_starting
- apache_scboard_reading
- apache_scboard_sending
- apache_scboard_keepalive
- apache_scboard_dnslookup
- apache_scboard_closing
- apache_scboard_logging
- apache_scboard_finishing
- apache_scboard_idle_cleanup
- apache_scboard_open
