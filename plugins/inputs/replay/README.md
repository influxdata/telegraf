# Replay Input Plugin

The replay plugin replays the data from a file following the original order and cadence,
using the selected [input data format](/docs/DATA_FORMATS_INPUT.md).


### Configuration:

```toml
[[inputs.replay]]
  ## Files to parse each interval.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/data/**.csv     -> recursively find all csv files in /var/data
  ##   /var/data/*/*.csv    -> find all .csv files with a parent dir in /var/data
  ##   /var/data/replay.csv -> only replay "replay.csv"
  files = ["/var/data/**.csv"]

  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "csv"

  ## CSV Specific configuration options
  csv_header_row_count = 1
  csv_timestamp_column = "time"
  csv_timestamp_format = "unix_ns"
  csv_measurement_column = "name"
  csv_trim_space = true
  
  ## How many times to iterate through the file. -1 to continually replay the 
  ## file over and over again
  iterations = -1

  ## Might be useful if the csv file has more data than you need
  # fielddrop = [ "column1", "column2" ]

  ## Name a tag containing the name of the file the data was parsed from.  Leave empty
  ## to disable.
  # file_tag = ""
```
