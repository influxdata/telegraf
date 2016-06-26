# file Output Plugin

This plugin writes to a file on disk.

# Configuration for output

## Files to write to, "stdout" is a specially handled file.
## For files which contain curly bracket tokens, these tokens will be interpretted as a date/time format
## so file will be generated based on provided format and UTC time on creation.
## This can be used to create dated directories or include time in name
## for example to create a file called metrics.out in a dir within /tmp with todays date use /tmp/{020106}/metric.out
## similarly if the filename was to also contain the current date and time on creation use /tmp/{020106}/metrics{020106.150406}.out
## for more info on token time format notation see https://golang.org/pkg/time/#Time.Format
files = ["stdout", "/tmp/metrics.out", "/tmp/{020106}/metrics{020106.150406}.out"]


## Data format to output.
## Each data format has it's own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
data_format = "influx"
