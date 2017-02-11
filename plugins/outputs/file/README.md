# file Output Plugin
This plugin writes to a file on disk.

# Configuration for output
Files to write to, "stdout" is a specially handled file.
For file names which contain curly bracket tokens, these tokens will be 
interpretted as a date/time format,so file will be generated based on provided
format and UTC time on creation.
This can be used to create dated directories or include time in name.
To create a file called data.out in a dir within /tmp with todays date use
/tmp/{020106}/data.out
To create a file which contains the current date and time in the name use 
/tmp/{020106}/data{020106.150406}.out
for more info on token format see https://golang.org/pkg/time/#Time.Format
files = ["stdout", "/tmp/data.out", "/tmp/{020106}/data{020106.150406}.out"]

## Data format to output
Each data format has it's own unique set of configuration options, read
more about them at: 
https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
data_format = "influx"