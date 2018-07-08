# Tail Input Plugin

The tail plugin "tails" a logfile and parses each log message.

By default, the tail plugin acts like the following unix tail command:

```
tail -F --lines=0 myfile.log
```

- `-F` means that it will follow the _name_ of the given file, so
that it will be compatible with log-rotated files, and that it will retry on
inaccessible files.
- `--lines=0` means that it will start at the end of the file (unless
the `from_beginning` option is set).

see http://man7.org/linux/man-pages/man1/tail.1.html for more details.

The plugin expects messages in one of the
[Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

### Configuration:

```toml
# Stream a log file, like the tail -f command
[[inputs.tail]]
  ## files to tail.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   "/var/log/**.log"  -> recursively find all .log files in /var/log
  ##   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
  ##   "/var/log/apache.log" -> just tail the apache log file
  ##
  ## See https://github.com/gobwas/glob for more examples
  ##
  files = ["/var/mymetrics.out"]
  ## Read file from beginning.
  from_beginning = false
  ## Whether file is a named pipe
  pipe = false

  ## Method used to watch for file updates.  Can be either "inotify" or "poll".
  # watch_method = "inotify"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Metrics:

Metrics are produced according to the `data_format` option.  Additionally a
tag labeled `path` is added to the metric containing the filename being tailed.
