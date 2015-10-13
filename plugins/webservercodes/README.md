A plugin that collects and counts http codes from any webserver that has access log.
It parses last several entries of log files and counts http codes.

Configuration.

A plugin supports per-virtualhost configuration. Config may contains one or more of this sections (example):

[[webservercodes.vhosts]]

host = "defaulthost"

access_log = "/var/log/apache2/access.log"

regex_parsestring = `\[(?P<time>[^\]]+)\] ".*?" (?P<code>\d{3})`

parse_interval = "10s"

'host' is the name of virtualhost. It's preferred to use main domain as a value.

'access_log' is the full name of log file for this virtualhost (with full path). File must exists ann be readable by telegraf process owner (usually user 'telegraf').

If you have apache web server installed on Debian/Ubuntu, you can set file readable by executing shell command:

`$ sudo adduser telegraf adm`

'regex_parsestring' is regexp pattern for log parsing. Regexp syntax must conform to goland acceptable syntax (see https://golang.org/s/re2syntax).

Pattern must include named substrings with labels 'time' and 'code'.

Example for apache "common" and "combined" log formats, nginx default log format ("combined"):

`\[(?P<time>[^\]]+)\] ".*?" (?P<code>\d{3})`

This pattern matches for log entries like (example):

`127.0.0.1 - - [30/Aug/2015:05:59:36 +0000] "GET / HTTP/1.1" 404 379 "-" "-"`

If you use another web server, please note that time must conform to apache2 `%t` format, see default configuration and http://httpd.apache.org/docs/2.2/mod/mod_log_config.html

'parse_interval' is interval that limits amount of log entries readed and parsed by plugin. It allows to examine even extremely large log files. 

Usually this setting may match to global 'interval' parameter, but it can be adjusted in some cases of too high or too low load.

Parameter must be in time.Duration format (see https://golang.org/pkg/time/#ParseDuration).

Collected measurements.

Plugin adds tag 'virtualhost' to standard 'host' tag to all it's output values.

Output values are:
- 'webservercodes_NNN' = (int), where NNN is all http codes defined in https://www.ietf.org/rfc/rfc2616.txt;
- 'webservercodes_total' = (int) - total amount of codes collected per one plugin start.
