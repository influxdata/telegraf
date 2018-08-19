# Logparser Input Plugin

### **Deprecated in version 1.8**: Please use the [tail](/plugins/inputs/tail) plugin with the `grok` [data format](/docs/DATA_FORMATS_INPUT.md).

The `logparser` plugin streams and parses the given logfiles. Currently it
has the capability of parsing "grok" patterns from logfiles, which also supports
regex patterns.

### Configuration:

```toml
[[inputs.logparser]]
  ## DEPRECATED: The `logparser` plugin is deprecated in 1.8.  Please use the
  ## `tail` plugin with the grok data_format instead.

  ## Log files to parse.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/log/**.log     -> recursively find all .log files in /var/log
  ##   /var/log/*/*.log    -> find all .log files with a parent dir in /var/log
  ##   /var/log/apache.log -> only tail the apache log file
  files = ["/var/log/apache/access.log"]

  ## Read files that currently exist from the beginning. Files that are created
  ## while telegraf is running (and that match the "files" globs) will always
  ## be read from the beginning.
  from_beginning = false

  ## Method used to watch for file updates.  Can be either "inotify" or "poll".
  # watch_method = "inotify"

  ## Parse logstash-style "grok" patterns:
  ##   Telegraf built-in parsing patterns: https://goo.gl/dkay10
  [inputs.logparser.grok]
    ## This is a list of patterns to check the given log file(s) for.
    ## Note that adding patterns here increases processing time. The most
    ## efficient configuration is to have one pattern per logparser.
    ## Other common built-in patterns are:
    ##   %{COMMON_LOG_FORMAT}   (plain apache & nginx access logs)
    ##   %{COMBINED_LOG_FORMAT} (access logs + referrer & agent)
    patterns = ["%{COMBINED_LOG_FORMAT}"]

    ## Name of the outputted measurement name.
    measurement = "apache_access_log"

    ## Full path(s) to custom pattern files.
    custom_pattern_files = []

    ## Custom patterns can also be defined here. Put one pattern per line.
    custom_patterns = '''
    '''

    ## Timezone allows you to provide an override for timestamps that
    ## don't already include an offset
    ## e.g. 04/06/2016 12:41:45 data one two 5.43µs
    ##
    ## Default: "" which renders UTC
    ## Options are as follows:
    ##   1. Local             -- interpret based on machine localtime
    ##   2. "Canada/Eastern"  -- Unix TZ values like those found in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
    ##   3. UTC               -- or blank/unspecified, will return timestamp in UTC
    timezone = "Canada/Eastern"
```

### Grok Parser

The best way to get acquainted with grok patterns is to read the logstash docs,
which are available here:
  https://www.elastic.co/guide/en/logstash/current/plugins-filters-grok.html

The Telegraf grok parser uses a slightly modified version of logstash "grok"
patterns, with the format

```
%{<capture_syntax>[:<semantic_name>][:<modifier>]}
```

The `capture_syntax` defines the grok pattern that's used to parse the input
line and the `semantic_name` is used to name the field or tag.  The extension
`modifier` controls the data type that the parsed item is converted to or
other special handling.

By default all named captures are converted into string fields.
Timestamp modifiers can be used to convert captures to the timestamp of the
parsed metric.  If no timestamp is parsed the metric will be created using the
current time.

You must capture at least one field per line.

- Available modifiers:
  - string   (default if nothing is specified)
  - int
  - float
  - duration (ie, 5.23ms gets converted to int nanoseconds)
  - tag      (converts the field into a tag)
  - drop     (drops the field completely)
- Timestamp modifiers:
  - ts               (This will auto-learn the timestamp format)
  - ts-ansic         ("Mon Jan _2 15:04:05 2006")
  - ts-unix          ("Mon Jan _2 15:04:05 MST 2006")
  - ts-ruby          ("Mon Jan 02 15:04:05 -0700 2006")
  - ts-rfc822        ("02 Jan 06 15:04 MST")
  - ts-rfc822z       ("02 Jan 06 15:04 -0700")
  - ts-rfc850        ("Monday, 02-Jan-06 15:04:05 MST")
  - ts-rfc1123       ("Mon, 02 Jan 2006 15:04:05 MST")
  - ts-rfc1123z      ("Mon, 02 Jan 2006 15:04:05 -0700")
  - ts-rfc3339       ("2006-01-02T15:04:05Z07:00")
  - ts-rfc3339nano   ("2006-01-02T15:04:05.999999999Z07:00")
  - ts-httpd         ("02/Jan/2006:15:04:05 -0700")
  - ts-epoch         (seconds since unix epoch, may contain decimal)
  - ts-epochnano     (nanoseconds since unix epoch)
  - ts-syslog        ("Jan 02 15:04:05", parsed time is set to the current year)
  - ts-"CUSTOM"

CUSTOM time layouts must be within quotes and be the representation of the
"reference time", which is `Mon Jan 2 15:04:05 -0700 MST 2006`.  
To match a comma decimal point you can use a period.  For example `%{TIMESTAMP:timestamp:ts-"2006-01-02 15:04:05.000"}` can be used to match `"2018-01-02 15:04:05,000"`
To match a comma decimal point you can use a period in the pattern string.
See https://golang.org/pkg/time/#Parse for more details.

Telegraf has many of its own [built-in patterns](./grok/patterns/influx-patterns),
as well as support for most of
[logstash's builtin patterns](https://github.com/logstash-plugins/logstash-patterns-core/blob/master/patterns/grok-patterns).
_Golang regular expressions do not support lookahead or lookbehind.
logstash patterns that depend on these are not supported._

If you need help building patterns to match your logs,
you will find the https://grokdebug.herokuapp.com application quite useful!

#### Timestamp Examples

This example input and config parses a file using a custom timestamp conversion:

```
2017-02-21 13:10:34 value=42
```

```toml
[[inputs.logparser]]
  [inputs.logparser.grok]
    patterns = ['%{TIMESTAMP_ISO8601:timestamp:ts-"2006-01-02 15:04:05"} value=%{NUMBER:value:int}']
```

This example input and config parses a file using a timestamp in unix time:

```
1466004605 value=42
1466004605.123456789 value=42
```

```toml
[[inputs.logparser]]
  [inputs.logparser.grok]
    patterns = ['%{NUMBER:timestamp:ts-epoch} value=%{NUMBER:value:int}']
```

This example parses a file using a built-in conversion and a custom pattern:

```
Wed Apr 12 13:10:34 PST 2017 value=42
```

```toml
[[inputs.logparser]]
  [inputs.logparser.grok]
	patterns = ["%{TS_UNIX:timestamp:ts-unix} value=%{NUMBER:value:int}"]
    custom_patterns = '''
      TS_UNIX %{DAY} %{MONTH} %{MONTHDAY} %{HOUR}:%{MINUTE}:%{SECOND} %{TZ} %{YEAR}
    '''
```

For cases where the timestamp itself is without offset, the `timezone` config var is available
to denote an offset. By default (with `timezone` either omit, blank or set to `"UTC"`), the times
are processed as if in the UTC timezone. If specified as `timezone = "Local"`, the timestamp
will be processed based on the current machine timezone configuration. Lastly, if using a
timezone from the list of Unix [timezones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones), the logparser grok will attempt to offset
the timestamp accordingly. See test cases for more detailed examples.

#### TOML Escaping

When saving patterns to the configuration file, keep in mind the different TOML
[string](https://github.com/toml-lang/toml#string) types and the escaping
rules for each.  These escaping rules must be applied in addition to the
escaping required by the grok syntax.  Using the Multi-line line literal
syntax with `'''` may be useful.

The following config examples will parse this input file:

```
|42|\uD83D\uDC2F|'telegraf'|
```

Since `|` is a special character in the grok language, we must escape it to
get a literal `|`.  With a basic TOML string, special characters such as
backslash must be escaped, requiring us to escape the backslash a second time.

```toml
[[inputs.logparser]]
  [inputs.logparser.grok]
    patterns = ["\\|%{NUMBER:value:int}\\|%{UNICODE_ESCAPE:escape}\\|'%{WORD:name}'\\|"]
    custom_patterns = "UNICODE_ESCAPE (?:\\\\u[0-9A-F]{4})+"
```

We cannot use a literal TOML string for the pattern, because we cannot match a
`'` within it.  However, it works well for the custom pattern.
```toml
[[inputs.logparser]]
  [inputs.logparser.grok]
    patterns = ["\\|%{NUMBER:value:int}\\|%{UNICODE_ESCAPE:escape}\\|'%{WORD:name}'\\|"]
    custom_patterns = 'UNICODE_ESCAPE (?:\\u[0-9A-F]{4})+'
```

A multi-line literal string allows us to encode the pattern:
```toml
[[inputs.logparser]]
  [inputs.logparser.grok]
    patterns = ['''
	  \|%{NUMBER:value:int}\|%{UNICODE_ESCAPE:escape}\|'%{WORD:name}'\|
	''']
    custom_patterns = 'UNICODE_ESCAPE (?:\\u[0-9A-F]{4})+'
```

#### Parsing Telegraf log file
We can use logparser to convert the log lines generated by Telegraf in metrics.

To do this we need to configure Telegraf to write logs to a file.
This could be done using the ``agent.logfile`` parameter or configuring syslog.
```toml
[agent]
  logfile = "/var/log/telegraf/telegraf.log"
```

Logparser configuration:
```toml
[[inputs.logparser]]
  files = ["/var/log/telegraf/telegraf.log"]

  [inputs.logparser.grok]
    measurement = "telegraf_log"
    patterns = ['^%{TIMESTAMP_ISO8601:timestamp:ts-rfc3339} %{TELEGRAF_LOG_LEVEL:level:tag}! %{GREEDYDATA:msg}']
    custom_patterns = '''
TELEGRAF_LOG_LEVEL (?:[DIWE]+)
'''
```

Example log lines:
```
2018-06-14T06:41:35Z I! Starting Telegraf v1.6.4
2018-06-14T06:41:35Z I! Agent Config: Interval:3s, Quiet:false, Hostname:"archer", Flush Interval:3s
2018-02-20T22:39:20Z E! Error in plugin [inputs.docker]: took longer to collect than collection interval (10s)
2018-06-01T10:34:05Z W! Skipping a scheduled flush because there is already a flush ongoing.
2018-06-14T07:33:33Z D! Output [file] buffer fullness: 0 / 10000 metrics.
```

Generated metrics:
```
telegraf_log,host=somehostname,level=I msg="Starting Telegraf v1.6.4" 1528958495000000000
telegraf_log,host=somehostname,level=I msg="Agent Config: Interval:3s, Quiet:false, Hostname:\"somehostname\", Flush Interval:3s" 1528958495001000000
telegraf_log,host=somehostname,level=E msg="Error in plugin [inputs.docker]: took longer to collect than collection interval (10s)" 1519166360000000000
telegraf_log,host=somehostname,level=W msg="Skipping a scheduled flush because there is already a flush ongoing." 1527849245000000000
telegraf_log,host=somehostname,level=D msg="Output [file] buffer fullness: 0 / 10000 metrics." 1528961613000000000
```


### Tips for creating patterns

Writing complex patterns can be difficult, here is some advice for writing a
new pattern or testing a pattern developed [online](https://grokdebug.herokuapp.com).

Create a file output that writes to stdout, and disable other outputs while
testing.  This will allow you to see the captured metrics.  Keep in mind that
the file output will only print once per `flush_interval`.

```toml
[[outputs.file]]
  files = ["stdout"]
```

- Start with a file containing only a single line of your input.
- Remove all but the first token or piece of the line.
- Add the section of your pattern to match this piece to your configuration file.
- Verify that the metric is parsed successfully by running Telegraf.
- If successful, add the next token, update the pattern and retest.
- Continue one token at a time until the entire line is successfully parsed.

### Additional Resources

- https://www.influxdata.com/telegraf-correlate-log-metrics-data-performance-bottlenecks/
