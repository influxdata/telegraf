# tail Input Plugin

The tail plugin "tails" a logfile and parses each log message.

By default, the tail plugin acts like the following unix tail command:

```
tail --follow=name --lines=0 --retry myfile.log
```

- `--follow=name` means that it will follow the _name_ of the given file, so
that it will be compatible with log-rotated files.
- `--lines=0` means that it will start at the end of the file (unless
the `from_beginning` option is set).
- `--retry` means it will retry on inaccessible files.

see http://man7.org/linux/man-pages/man1/tail.1.html for more details.

The plugin expects messages in one of the
[Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

### Configuration:

```toml
# Stream a log file, like the tail -f command
[[inputs.tail]]
  # SampleConfig
```

