# Directory Monitor Input Plugin

This plugin monitors a single directory (without looking at sub-directories), and takes in each file placed in the directory.
The plugin will gather all files in the directory at a configurable interval (`monitor_interval`), and parse the ones that haven't been picked up yet.

Please be advised that this plugin pulls files directly after they've been in the directory for the length of the configurable `directory_duration_threshold`, and thus files should not be written 'live' to the monitored directory unless they are guaranteed to finish writing before the `directory_duration_threshold`. This plugin is intended to read files that are moved or copied to the monitored directory, and thus files should also not be used by another process or else they may fail to be gathered.
This plugin is intended to read files that are moved or copied to the monitored directory, and thus files should also not be used by another process or else they may fail to be gathered. Please be advised that this plugin pulls files directly after they've been in the directory for the length of the configurable `directory_duration_threshold`, and thus files should not be written 'live' to the monitored directory. If you absolutely must write files directly, they must be guaranteed to finish writing before the `directory_duration_threshold`.

### Configuration:

```toml
## The directory to monitor and read files from.
directory = ""
#
## The directory to move finished files to.
finished_directory = "5s"
#
## Whether or not to move files that error out to an error directory.
use_error_directory = "true"
#
## The directory to move files to upon file error, given that 'use_error_directory' is enabled.
## If not is given, the error directory will be auto-generated.
# error_directory = ""
#
## The interval at which to check the directory for new files.
# monitor_interval = "50ms"
#
## The amount of time a file is allowed to sit in the directory before it is picked up.
## This time can generally be low but if you choose to have a very large file written to the directory and it's potentially slow,
## set this higher so that the plugin will wait until the file is fully copied to the directory.
# directory_duration_threshold = "50ms"
#
## Character encoding to use when interpreting the file contents. Invalid
## characters are replaced using the unicode replacement character. Defaults to utf-8.
##   ex: character_encoding = "utf-8"
##       character_encoding = "utf-16le"
##       character_encoding = "utf-16be"
# character_encoding = "utf-8"
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
## The maximum amount of files to process at once. A very high number can lead to bigger memory use and potential file system errors.
# max_concurrent_files = 3000
#
## The dataformat to be read from the files.
## Each data format has its own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
## NOTE: We do not currently support JSON file streaming and thus JSON files will be fully loaded into memory when they are processed.
data_format = "influx"
```
