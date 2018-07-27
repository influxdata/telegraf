# File Input Plugin

The file plugin updates a list of files every interval and parses the contents
using the selected [input data format](/docs/DATA_FORMATS_INPUT.md).

Files will always be read in their entirety, if you wish to tail/follow a file
use the [tail input plugin](/plugins/inputs/tail) instead.

### Configuration:
```toml
[[inputs.file]]
  ## Files to parse each interval.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/log/**.log     -> recursively find all .log files in /var/log
  ##   /var/log/*/*.log    -> find all .log files with a parent dir in /var/log
  ##   /var/log/apache.log -> only read the apache log file
  files = ["/var/log/apache/access.log"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```
