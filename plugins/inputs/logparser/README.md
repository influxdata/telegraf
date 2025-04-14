# Logparser Input Plugin

This service plugin streams and parses the given logfiles. Currently it
has the capability of parsing "grok" patterns from logfiles, which also supports
regex patterns.

> [!IMPORTANT]
> This plugin is deprecated. Please use the [`tail` plugin][tail] plugin in
> combination with the [`grok` data format][grok_parser] as a replacement.

⭐ Telegraf v1.0.0
🚩 Telegraf v1.15.0
🔥 Telegraf v1.35.0
🏷️ system, logging
💻 freebsd, linux, macos, windows

## Migration guide

This plugin is deprecated since Telegraf v1.15. To replace the plugin please
use the [`tail` plugin][tail] plugin in combination with the
[`grok` data format][grok_parser].

Here an example for replacing the existing instance:

```diff
- [[inputs.logparser]]
-   files = ["/var/log/apache/access.log"]
-   from_beginning = false
-   [inputs.logparser.grok]
-     patterns = ["%{COMBINED_LOG_FORMAT}"]
-     measurement = "apache_access_log"
-     custom_pattern_files = []
-     custom_patterns = '''
-     '''
-     timezone = "Canada/Eastern"

+ [[inputs.tail]]
+   files = ["/var/log/apache/access.log"]
+   from_beginning = false
+   grok_patterns = ["%{COMBINED_LOG_FORMAT}"]
+   name_override = "apache_access_log"
+   grok_custom_pattern_files = []
+   grok_custom_patterns = '''
+   '''
+   grok_timezone = "Canada/Eastern"
+   data_format = "grok"
```

[tail]: /plugins/inputs/tail/README.md
[grok_parser]: /plugins/parsers/grok/README.md

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
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
# Read metrics off Arista LANZ, via socket
[[inputs.logparser]]
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
    # timezone = "Canada/Eastern"

    ## When set to "disable", timestamp will not incremented if there is a
    ## duplicate.
    # unique_timestamp = "auto"
```

## Metrics

The plugin accepts arbitrary input and parses it according to the `grok`
patterns configured. There is no predefined metric format.

## Example Output

There is no predefined metric format, so output depends on plugin input.
