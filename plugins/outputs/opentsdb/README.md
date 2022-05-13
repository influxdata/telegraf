# OpenTSDB Output Plugin

This plugin writes to an OpenTSDB instance using either the "telnet" or Http
mode.

Using the Http API is the recommended way of writing metrics since OpenTSDB 2.0
To use Http mode, set useHttp to true in config. You can also control how many
metrics is sent in each http request by setting batchSize in config.

See [the docs](http://opentsdb.net/docs/build/html/api_http/put.html) for
details.

## Configuration

```toml
# Configuration for OpenTSDB server to send metrics to
[[outputs.opentsdb]]
  ## prefix for metrics keys
  prefix = "my.specific.prefix."

  ## DNS name of the OpenTSDB server
  ## Using "opentsdb.example.com" or "tcp://opentsdb.example.com" will use the
  ## telnet API. "http://opentsdb.example.com" will use the Http API.
  host = "opentsdb.example.com"

  ## Port of the OpenTSDB server
  port = 4242

  ## Number of data points to send to OpenTSDB in Http requests.
  ## Not used with telnet API.
  http_batch_size = 50

  ## URI Path for Http requests to OpenTSDB.
  ## Used in cases where OpenTSDB is located behind a reverse proxy.
  http_path = "/api/put"

  ## Debug true - Prints OpenTSDB communication
  debug = false

  ## Separator separates measurement name from field
  separator = "_"
```

## Transfer "Protocol" in the telnet mode

The expected input from OpenTSDB is specified in the following way:

```text
put <metric> <timestamp> <value> <tagk1=tagv1[ tagk2=tagv2 ...tagkN=tagvN]>
```

The telegraf output plugin adds an optional prefix to the metric keys so that a
subamount can be selected.

```text
put <[prefix.]metric> <timestamp> <value> <tagk1=tagv1[ tagk2=tagv2 ...tagkN=tagvN]>
```

### Example

```text
put nine.telegraf.system_load1 1441910356 0.430000 dc=homeoffice host=irimame scope=green
put nine.telegraf.system_load5 1441910356 0.580000 dc=homeoffice host=irimame scope=green
put nine.telegraf.system_load15 1441910356 0.730000 dc=homeoffice host=irimame scope=green
put nine.telegraf.system_uptime 1441910356 3655970.000000 dc=homeoffice host=irimame scope=green
put nine.telegraf.system_uptime_format 1441910356  dc=homeoffice host=irimame scope=green
put nine.telegraf.mem_total 1441910356 4145426432 dc=homeoffice host=irimame scope=green
...
put nine.telegraf.io_write_bytes 1441910366 0 dc=homeoffice host=irimame name=vda2 scope=green
put nine.telegraf.io_read_time 1441910366 0 dc=homeoffice host=irimame name=vda2 scope=green
put nine.telegraf.io_write_time 1441910366 0 dc=homeoffice host=irimame name=vda2 scope=green
put nine.telegraf.io_io_time 1441910366 0 dc=homeoffice host=irimame name=vda2 scope=green
put nine.telegraf.ping_packets_transmitted 1441910366  dc=homeoffice host=irimame scope=green url=www.google.com
put nine.telegraf.ping_packets_received 1441910366  dc=homeoffice host=irimame scope=green url=www.google.com
put nine.telegraf.ping_percent_packet_loss 1441910366 0.000000 dc=homeoffice host=irimame scope=green url=www.google.com
put nine.telegraf.ping_average_response_ms 1441910366 24.006000 dc=homeoffice host=irimame scope=green url=www.google.com
...
```

The OpenTSDB telnet interface can be simulated with this reader:

```go
// opentsdb_telnet_mode_mock.go
package main

import (
    "io"
    "log"
    "net"
    "os"
)

func main() {
    l, err := net.Listen("tcp", "localhost:4242")
    if err != nil {
        log.Fatal(err)
    }
    defer l.Close()
    for {
        conn, err := l.Accept()
        if err != nil {
            log.Fatal(err)
        }
        go func(c net.Conn) {
            defer c.Close()
            io.Copy(os.Stdout, c)
        }(conn)
    }
}

```

## Allowed values for metrics

OpenTSDB allows `integers` and `floats` as input values
