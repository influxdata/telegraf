# Execd Processor Plugin

The `execd` processor plugin runs an external program as a separate process and
pipes metrics in to the process's STDIN and reads processed metrics from its STDOUT.
The programs must accept influx line protocol on standard in (STDIN) and output
metrics in influx line protocol to standard output (STDOUT).

Program output on standard error is mirrored to the telegraf log.

Telegraf minimum version: Telegraf 1.15.0

## Caveats

- Metrics with tracking will be considered "delivered" as soon as they are passed
  to the external process. There is currently no way to match up which metric
  coming out of the execd process relates to which metric going in (keep in mind
  that processors can add and drop metrics, and that this is all done
  asynchronously).
- it's not currently possible to use a data_format other than "influx", due to
  the requirement that it is serialize-parse symmetrical and does not lose any
  critical type data.

## Configuration

```toml
# Run executable as long-running processor plugin
[[processors.execd]]
  ## One program to run as daemon.
  ## NOTE: process and each argument should each be their own string
  ## eg: command = ["/path/to/your_program", "arg1", "arg2"]
  command = ["cat"]

  ## Delay before the process is restarted after an unexpected termination
  # restart_delay = "10s"
```

## Example

### Go daemon example

This go daemon reads a metric from stdin, multiplies the "count" field by 2,
and writes the metric back out.

```go
package main

import (
    "fmt"
    "os"

    "github.com/influxdata/telegraf/metric"
    "github.com/influxdata/telegraf/plugins/parsers/influx"
    "github.com/influxdata/telegraf/plugins/serializers"
)

func main() {
    parser := influx.NewStreamParser(os.Stdin)
    serializer, _ := serializers.NewInfluxSerializer()

    for {
        metric, err := parser.Next()
        if err != nil {
            if err == influx.EOF {
                return // stream ended
            }
            if parseErr, isParseError := err.(*influx.ParseError); isParseError {
                fmt.Fprintf(os.Stderr, "parse ERR %v\n", parseErr)
                os.Exit(1)
            }
            fmt.Fprintf(os.Stderr, "ERR %v\n", err)
            os.Exit(1)
        }

        c, found := metric.GetField("count")
        if !found {
            fmt.Fprintf(os.Stderr, "metric has no count field\n")
            os.Exit(1)
        }
        switch t := c.(type) {
        case float64:
            t *= 2
            metric.AddField("count", t)
        case int64:
            t *= 2
            metric.AddField("count", t)
        default:
            fmt.Fprintf(os.Stderr, "count is not an unknown type, it's a %T\n", c)
            os.Exit(1)
        }
        b, err := serializer.Serialize(metric)
        if err != nil {
            fmt.Fprintf(os.Stderr, "ERR %v\n", err)
            os.Exit(1)
        }
        fmt.Fprint(os.Stdout, string(b))
    }
}
```

to run it, you'd build the binary using go, eg `go build -o multiplier.exe main.go`

```toml
[[processors.execd]]
  command = ["multiplier.exe"]
```

### Ruby daemon

- See [Ruby daemon](./examples/multiplier_line_protocol/multiplier_line_protocol.rb)

```toml
[[processors.execd]]
  command = ["ruby", "plugins/processors/execd/examples/multiplier_line_protocol/multiplier_line_protocol.rb"]
```
