package grok

const DEFAULT_PATTERNS = `
# Captures are a slightly modified version of logstash "grok" patterns, with
#  the format %{<capture syntax>[:<semantic name>][:<modifier>]}
# By default all named captures are converted into string fields.
# If a pattern does not have a semantic name it will not be captured.
# Modifiers can be used to convert captures to other types or tags.
# Timestamp modifiers can be used to convert captures to the timestamp of the
#  parsed metric.

# View logstash grok pattern docs here:
#   https://www.elastic.co/guide/en/logstash/current/plugins-filters-grok.html
# All default logstash patterns are supported, these can be viewed here:
#   https://github.com/logstash-plugins/logstash-patterns-core/blob/master/patterns/grok-patterns

# Available modifiers:
#   string   (default if nothing is specified)
#   int
#   float
#   duration (ie, 5.23ms gets converted to int nanoseconds)
#   tag      (converts the field into a tag)
#   drop     (drops the field completely)
# Timestamp modifiers:
#   ts-ansic         ("Mon Jan _2 15:04:05 2006")
#   ts-unix          ("Mon Jan _2 15:04:05 MST 2006")
#   ts-ruby          ("Mon Jan 02 15:04:05 -0700 2006")
#   ts-rfc822        ("02 Jan 06 15:04 MST")
#   ts-rfc822z       ("02 Jan 06 15:04 -0700")
#   ts-rfc850        ("Monday, 02-Jan-06 15:04:05 MST")
#   ts-rfc1123       ("Mon, 02 Jan 2006 15:04:05 MST")
#   ts-rfc1123z      ("Mon, 02 Jan 2006 15:04:05 -0700")
#   ts-rfc3339       ("2006-01-02T15:04:05Z07:00")
#   ts-rfc3339nano   ("2006-01-02T15:04:05.999999999Z07:00")
#   ts-httpd         ("02/Jan/2006:15:04:05 -0700")
#   ts-epoch         (seconds since unix epoch)
#   ts-epochnano     (nanoseconds since unix epoch)
#   ts-"CUSTOM"
# CUSTOM time layouts must be within quotes and be the representation of the
# "reference time", which is Mon Jan 2 15:04:05 -0700 MST 2006
# See https://golang.org/pkg/time/#Parse for more details.

# Example log file pattern, example log looks like this:
#   [04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs
# Breakdown of the DURATION pattern below:
#   NUMBER  is a builtin logstash grok pattern matching float & int numbers.
#   [nuµm]? is a regex specifying 0 or 1 of the characters within brackets.
#   s       is also regex, this pattern must end in "s".
# so DURATION will match something like '5.324ms' or '6.1µs' or '10s'
DURATION %{NUMBER}[nuµm]?s
RESPONSE_CODE %{NUMBER:response_code:tag}
RESPONSE_TIME %{DURATION:response_time_ns:duration}
EXAMPLE_LOG \[%{HTTPDATE:ts:ts-httpd}\] %{NUMBER:myfloat:float} %{RESPONSE_CODE} %{IPORHOST:clientip} %{RESPONSE_TIME}

# Wider-ranging username matching vs. logstash built-in %{USER}
NGUSERNAME [a-zA-Z0-9\.\@\-\+_%]+
NGUSER %{NGUSERNAME}
# Wider-ranging client IP matching
CLIENT (?:%{IPV6}|%{IPV4}|%{HOSTNAME}|%{HOSTPORT})

##
## COMMON LOG PATTERNS
##

# apache & nginx logs, this is also known as the "common log format"
#   see https://en.wikipedia.org/wiki/Common_Log_Format
COMMON_LOG_FORMAT %{CLIENT:client_ip} %{NOTSPACE:ident} %{NOTSPACE:auth} \[%{HTTPDATE:ts:ts-httpd}\] "(?:%{WORD:verb:tag} %{NOTSPACE:request}(?: HTTP/%{NUMBER:http_version:float})?|%{DATA})" %{NUMBER:resp_code:tag} (?:%{NUMBER:resp_bytes:int}|-)

# Combined log format is the same as the common log format but with the addition
# of two quoted strings at the end for "referrer" and "agent"
#   See Examples at http://httpd.apache.org/docs/current/mod/mod_log_config.html
COMBINED_LOG_FORMAT %{COMMON_LOG_FORMAT} "%{DATA:referrer}" "%{DATA:agent}"

# HTTPD log formats
HTTPD20_ERRORLOG \[%{HTTPDERROR_DATE:timestamp}\] \[%{LOGLEVEL:loglevel:tag}\] (?:\[client %{IPORHOST:clientip}\] ){0,1}%{GREEDYDATA:errormsg}
HTTPD24_ERRORLOG \[%{HTTPDERROR_DATE:timestamp}\] \[%{WORD:module}:%{LOGLEVEL:loglevel:tag}\] \[pid %{POSINT:pid:int}:tid %{NUMBER:tid:int}\]( \(%{POSINT:proxy_errorcode:int}\)%{DATA:proxy_errormessage}:)?( \[client %{IPORHOST:client}:%{POSINT:clientport}\])? %{DATA:errorcode}: %{GREEDYDATA:message}
HTTPD_ERRORLOG %{HTTPD20_ERRORLOG}|%{HTTPD24_ERRORLOG}
`
