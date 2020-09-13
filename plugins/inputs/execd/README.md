# Execd Input Plugin

The `execd` plugin runs an external program as a long-running daemon. 
The programs must output metrics in any one of the accepted 
[Input Data Formats][] on the process's STDOUT, and is expected to
stay running. If you'd instead like the process to collect metrics and then exit,
check out the [inputs.exec][] plugin.

The `signal` can be configured to send a signal the running daemon on each
collection interval. This is used for when you want to have Telegraf notify the
plugin when it's time to run collection. STDIN is recommended, which writes a
new line to the process's STDIN.

STDERR from the process will be relayed to Telegraf as errors in the logs.

STDIN (the daemon process's standard input) will be closed to indicate when
to shut down.  This can happen due to a normal telegraf shutdown, or when
telegraf's configuration has been refreshed via a SIGHUP.

### Configuration:

```toml
[[inputs.execd]]
  ## One program to run as daemon.
  ## NOTE: process and each argument should each be their own string
  command = ["telegraf-smartctl", "-d", "/dev/sda"]

  ## Define how the process is signaled on each collection interval.
  ## Valid values are:
  ##   "none"    : Do not signal anything. (Recommended for service inputs)
  ##               The process must output metrics by itself.
  ##   "STDIN"   : Send a newline on STDIN. (Recommended for gather inputs).
  ##               Note that it is standard input from the perspective of the
  #                daemon, not telegraf
  ##   "SIGHUP"  : Send a HUP signal. Not available on Windows. (not recommended)
  ##   "SIGUSR1" : Send a USR1 signal. Not available on Windows.
  ##   "SIGUSR2" : Send a USR2 signal. Not available on Windows.
  signal = "none"

  ## Startup string
  ##
  ## The standard way of configuring a daemon is through the command line.
  ## If the daemon requires additional configuratin, this is normally done
  ## through its own configuration file.  In some cases, the command line
  ## might provide the daemon with a path to such a configuration.
  ##
  ## As an alternative, some daemons may allow configuration through their
  ## standard input, as if a user had typed them or redirected them from
  ## a file.  Exed allows this by setting a write_on_start parameter.
  ## 
  ## !IMPORTANT! Newlines are not automatic - add \n if you want one.
  ## (You can also send a null character with \000)  
  ##
  ## write_on_start is sent ONCE when the daemon is started.  If this string is
  ## not specified or is zero-length, nothing is sent.
  ##
  ## !IMPORTANT! Newlines are not automatic - add \n if you want one.  
  ##
  ## write_on_start is sent ONCE when the daemon is started.  If this string is
  ## not specified or is zero-length, nothing is sent.
  write_on_start=""

  ## String to trigger periodic Gather (metric updates) from daemon
  ## !IMPORTANT! This setting is ignored unless signal="STDIN".
  ## !IMPORTANT! Newlines are not automatic - add \n if you want one.  
  ##
  ## write_on_start is sent ONCE when the daemon is started.  If this string is
  ## not specified or is zero-length, nothing is sent.
  ##
  write_on_start=""

  ## String to trigger periodic Gather (metric updates) from daemon
  ## String on gather
  ##
  ## !IMPORTANT! These settings are ignored unless signal="STDIN".
  ## !IMPORTANT! Newlines are not automatic - add \n if you want one.  
  ##
  ## write_on_gather is sent on every collection interval, as a mechanism for
  ## signaling the daemon that telegraf wants it to collect and transmit one or
  ## more metrics.  If signal="STDIN" but this string is not specified, then
  ## a single newline (\n) is used.  Also note that TOML accepts multi-line
  ## inputs.  Newlines are not automatic - end strings with \n if they are required.
  ##
  ## Note that regardless of this setting the daemon proesses's stdin is closed
  ## just prior to terminating the process (when telegraf shuts down)
  ##
  write_on_gather="\n"

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Example

##### Daemon written in bash using STDIN signaling

```bash
#!/bin/bash

counter=0

while IFS= read -r LINE; do
    echo "counter_bash count=${counter}"
    let counter=counter+1
done
```

```toml
[[inputs.execd]]
  command = ["plugins/inputs/execd/examples/count.sh"]
  signal = "STDIN"
```

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
[[inputs.execd]]
  command = ["plugins/inputs/execd/examples/count.go.exe"]
  signal = "SIGHUP"
```

##### Ruby daemon running standalone

```ruby
#!/usr/bin/env ruby

counter = 0

loop do
  puts "counter_ruby count=#{counter}"
  STDOUT.flush

  counter += 1
  sleep 1
end
```

```toml
[[inputs.execd]]
  command = ["plugins/inputs/execd/examples/count.rb"]
  signal = "none"
```

[Input Data Formats]: https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
[inputs.exec]: https://github.com/influxdata/telegraf/blob/master/plugins/inputs/exec/README.md
