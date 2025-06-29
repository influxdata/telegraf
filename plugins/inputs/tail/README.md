# Tail Input Plugin

This service plugin continuously reads a file and parses new data as it arrives
similar to the [tail -f command][tail]. The incoming messages are expected to be
in one of the supported [data formats][data_formats].

⭐ Telegraf v1.1.2
🏷️ logging
💻 all

[tail]: https://man7.org/linux/man-pages/man1/tail.1.html
[data_formats]: /docs/DATA_FORMATS_INPUT.md

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
# Parse the new lines appended to a file
[[inputs.tail]]
  ## File names or a pattern to tail.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   "/var/log/**.log"  -> recursively find all .log files in /var/log
  ##   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
  ##   "/var/log/apache.log" -> just tail the apache log file
  ##   "/var/log/log[!1-2]*  -> tail files without 1-2
  ##   "/var/log/log[^1-2]*  -> identical behavior as above
  ## See https://github.com/gobwas/glob for more examples
  ##
  files = ["/var/mymetrics.out"]

  ## Offset to start reading at
  ## The following methods are available:
  ##   beginning          -- start reading from the beginning of the file ignoring any persisted offset
  ##   end                -- start reading from the end of the file ignoring any persisted offset
  ##   saved-or-beginning -- use the persisted offset of the file or, if no offset persisted, start from the beginning of the file
  ##   saved-or-end       -- use the persisted offset of the file or, if no offset persisted, start from the end of the file
  # initial_read_offset = "saved-or-end"

  ## Whether file is a named pipe
  # pipe = false

  ## Method used to watch for file updates.  Can be either "inotify" or "poll".
  ## inotify is supported on linux, *bsd, and macOS, while Windows requires
  ## using poll. Poll checks for changes every 250ms.
  # watch_method = "inotify"

  ## Maximum lines of the file to process that have not yet be written by the
  ## output.  For best throughput set based on the number of metrics on each
  ## line and the size of the output's metric_batch_size.
  # max_undelivered_lines = 1000

  ## Character encoding to use when interpreting the file contents.  Invalid
  ## characters are replaced using the unicode replacement character.  When set
  ## to the empty string the data is not decoded to text.
  ##   ex: character_encoding = "utf-8"
  ##       character_encoding = "utf-16le"
  ##       character_encoding = "utf-16be"
  ##       character_encoding = ""
  # character_encoding = ""

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Set the tag that will contain the path of the tailed file. If you don't want this tag, set it to an empty string.
  # path_tag = "path"

  ## Filters to apply to files before generating metrics
  ## "ansi_color" removes ANSI colors
  # filters = []

  ## multiline parser/codec
  ## https://www.elastic.co/guide/en/logstash/2.4/plugins-filters-multiline.html
  #[inputs.tail.multiline]
    ## The pattern should be a regexp which matches what you believe to be an indicator that the field is part of an event consisting of multiple lines of log data.
    #pattern = "^\s"

    ## The field's value must be previous or next and indicates the relation to the
    ## multi-line event.
    #match_which_line = "previous"

    ## The invert_match can be true or false (defaults to false).
    ## If true, a message not matching the pattern will constitute a match of the multiline filter and the what will be applied. (vice-versa is also true)
    #invert_match = false

    ## The handling method for quoted text (defaults to 'ignore').
    ## The following methods are available:
    ##   ignore  -- do not consider quotation (default)
    ##   single-quotes -- consider text quoted by single quotes (')
    ##   double-quotes -- consider text quoted by double quotes (")
    ##   backticks     -- consider text quoted by backticks (`)
    ## When handling quotes, escaped quotes (e.g. \") are handled correctly.
    #quotation = "ignore"

    ## The preserve_newline option can be true or false (defaults to false).
    ## If true, the newline character is preserved for multiline elements,
    ## this is useful to preserve message-structure e.g. for logging outputs.
    #preserve_newline = false

    #After the specified timeout, this plugin sends the multiline event even if no new pattern is found to start a new event. The default is 5s.
    #timeout = 5s
```

## Metrics

Metrics are produced according to the `data_format` option.  Additionally a
tag labeled `path` is added to the metric containing the filename being tailed.

## Example Output

There is no predefined metric format, so output depends on plugin input.
