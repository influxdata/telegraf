# Directory Monitor Input Plugin

This plugin monitors a single directory (traversing sub-directories),
and takes in each file placed in the directory.  The plugin will gather all
files in the directory at the configured interval, and parse the ones that
haven't been picked up yet.

This plugin is intended to read files that are moved or copied to the monitored
directory, and thus files should also not be used by another process or else
they may fail to be gathered. Please be advised that this plugin pulls files
directly after they've been in the directory for the length of the configurable
`directory_duration_threshold`, and thus files should not be written 'live' to
the monitored directory. If you absolutely must write files directly, they must
be guaranteed to finish writing before the `directory_duration_threshold`.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Ingests files in a directory and then moves them to a target directory.
[[inputs.directory_monitor]]
  ## The directory to monitor and read files from (including sub-directories if "recursive" is true).
  directory = ""
  #
  ## The directory to move finished files to (maintaining directory hierachy from source).
  finished_directory = ""
  #
  ## Setting recursive to true will make the plugin recursively walk the directory and process all sub-directories.
  # recursive = false
  #
  ## The directory to move files to upon file error.
  ## If not provided, erroring files will stay in the monitored directory.
  # error_directory = ""
  #
  ## The amount of time a file is allowed to sit in the directory before it is picked up.
  ## This time can generally be low but if you choose to have a very large file written to the directory and it's potentially slow,
  ## set this higher so that the plugin will wait until the file is fully copied to the directory.
  # directory_duration_threshold = "50ms"
  #
  ## A list of the only file names to monitor, if necessary. Supports regex. If left blank, all files are ingested.
  # files_to_monitor = ["^.*\.csv"]
  #
  ## A list of files to ignore, if necessary. Supports regex.
  # files_to_ignore = [".DS_Store"]
  #
  ## Maximum lines of the file to process that have not yet be written by the
  ## output. For best throughput set to the size of the output's metric_buffer_limit.
  ## Warning: setting this number higher than the output's metric_buffer_limit can cause dropped metrics.
  # max_buffered_metrics = 10000
  #
  ## The maximum amount of file paths to queue up for processing at once, before waiting until files are processed to find more files.
  ## Lowering this value will result in *slightly* less memory use, with a potential sacrifice in speed efficiency, if absolutely necessary.
  # file_queue_size = 100000
  #
  ## Name a tag containing the name of the file the data was parsed from.  Leave empty
  ## to disable. Cautious when file name variation is high, this can increase the cardinality
  ## significantly. Read more about cardinality here:
  ## https://docs.influxdata.com/influxdb/cloud/reference/glossary/#series-cardinality
  # file_tag = ""
  #
  ## Specify if the file can be read completely at once or if it needs to be read line by line (default).
  ## Possible values: "line-by-line", "at-once"
  # parse_method = "line-by-line"
  #
  ## The dataformat to be read from the files.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## Metrics

The format of metrics produced by this plugin depends on the content and data
format of the file.

## Example Output

The metrics produced by this plugin depends on the content and data
format of the file.
