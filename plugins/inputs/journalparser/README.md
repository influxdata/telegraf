# journalparser input plugin

The journalparser input plugin streams from the systemd journal, parsing log entries into metrics. The plugin parses entries using "grok" patterns.

## Configuration

Configuration is similar to the [logparser](../logparser) plugin. For more detailed information on configuration of the grok parser, see the [logparser readme](../logparser/README.md#grok-parser).

### Journal fields
As opposed to normal log files, the systemd journal is not just a series of lines. The journal is a series of events, where each event has multiple fields. The journalparser plugin makes use of these fields in 2 places, `matches` and `grok.patterns`.

You can view these fields by using a command such as `journalctl -o verbose` or `journalctl -o json`.

### Example

```toml
[[inputs.journalparser]]
  ## Match filters to apply to journal entries. Only entries which match will be processed.
  matches = ["_COMM=httpd"]
  ## Read journal from beginning.
  from_beginning = false

  ## Parse logstash-style "grok" patterns:
  ##   Telegraf built-in parsing patterns: https://goo.gl/dkay10
  [inputs.journalparser.grok]
    ## Name of the outputted measurement name.
    measurement = "apache_access_log"
    ## Full path(s) to custom pattern files.
    custom_pattern_files = []
    ## Custom patterns can also be defined here. Put one pattern per line.
    custom_patterns = '''
    '''

    [inputs.journalparser.grok.patterns]
      ## This is a list of patterns to check the given log file(s) for.
      ## The parameter name is the journal field to apply the pattern on,
      ## typically "MESSAGE".
      ## Note that adding patterns here increases processing time. The most
      ## efficient configuration is to have one pattern per logparser.
      ## Other common built-in patterns are:
      ##   %{COMMON_LOG_FORMAT}   (plain apache & nginx access logs)
      ##   %{COMBINED_LOG_FORMAT} (access logs + referrer & agent)
      MESSAGE = ["%{COMBINED_LOG_FORMAT}"]
```
