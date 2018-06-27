# Reader Input Plugin

The Reader Plugin updates a list of files every interval and parses the data inside.
Files will always be read from the beginning.  
This plugin can parse any "data_format" formats.

### Configuration:
```toml
[[inputs.reader]]
## Files to parse each interval.
## These accept standard unix glob matching rules, but with the addition of
## ** as a "super asterisk". ie:
##   /var/log/**.log     -> recursively find all .log files in /var/log
##   /var/log/*/*.log    -> find all .log files with a parent dir in /var/log
##   /var/log/apache.log -> only tail the apache log file
files = ["/var/log/apache/access.log"]

## The dataformat to be read from files
## Each data format has its own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
data_format = ""
```