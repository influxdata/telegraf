# Execd Processor Plugin

The `execd` processor plugin runs an external program as a separate process and pipes metrics in to the process's STDIN and reads processed metrics from its STDOUT. 
The programs must output metrics in any one of the accepted 
[Processor Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md) on its standard output.

The `signal` can be configured to send a signal the running daemon on each collection interval.

Program output on standard error is mirrored to the telegraf log.

### Configuration:

```toml
[[processor.execd]]
  ## Program to run as daemon
  command = ["telegraf-smartctl", "-d", "/dev/sda"]

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"

  ## Data format your plugin will consume AND output. 
  ## Must be supported by both a serializer and a parser!
  ## 
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  ## 
  ## This should probably be either "json" or "influx"
  data_format = "json"
```

### Example

##### Go daemon using SIGHUP

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGHUP)

    counter := 0

    for {
        <-c

        fmt.Printf("counter_go count=%d\n", counter)
        counter++
    }
}

```

```toml
[[processors.execd]]
  command = ["plugins/processors/execd/examples/count.exe"]
```

- [Ruby daemon](./examples/multiplier.rb)

```toml
[[processors.execd]]
  command = ["plugins/processors/execd/examples/count.rb"]
```
