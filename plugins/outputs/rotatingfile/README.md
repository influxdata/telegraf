# rotatingfile Output Plugin
This plugin works exactly the same as the file output plugin, but the file is rotated. This practical if you for example use something that grabs those files and moves them across a network boundary or similar.

# Configuration
```
 [[outputs.rotating_file]]
   ## Files to write to, "stdout" is a specially handled file.
   root = "/tmp"
   filename_prefix = "metrics"
   max_age = "1m"

   ## Data format to output.
   ## Each data format has it's own unique set of configuration options, read
   ## more about them here:  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.m
   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
   data_format = "influx"
```
