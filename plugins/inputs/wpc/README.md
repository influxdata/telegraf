# wpc Input Plugin

This plugin gathers stats from [Windows Performance Counters](https://msdn.microsoft.com/en-us/library/windows/desktop/aa373083(v=vs.85).aspx)

### Configuration:

```toml
 # A plugin to collect stats from Windows Performance Counters
 [[inputs.wpc]]
  ## If the system being polled for data does not have a particular Counter at startup 
  ## of the Telegraf agent, it will not be gathered.
  # Prints all matching performance counters (useful for debugging)
  # PrintValid = false

  [[inputs.wpc.template]]
    # Processor usage, alternative to native.
    Counters = [
      # Use double-backslashes to work around a TOML parsing issue.
      [ "usage_idle", "\\Processor(*)\\%% Idle Time" ],
      [ "usage_user", "\\Processor(*)\\%% User Time" ],
      [ "usage_system", "\\Processor(*)\\%% Processor Time" ]
    ]
    Measurement = "win_cpu"
    # Print out when the performance counter is missing from object, counter or instance.
    # WarnOnMissing = false
```

## `wpc` vs `win_perf_counters`

The `win_perf_counters` plugin generates tags and fields using native Windows names.  This can make it difficult to compare common measurements across heterogenous environments because Windows names tend towards complexity.  For example, on Windows the performance counter "\\Processor(*)\\%% User Time" is equivalent to the Linux metric "cpu.usage_user" - good luck displaying both series on the same plot in Grafana.

Additionally, `win_perf_counters` can generate an large number of series in an InfluxDB database due to the inclusion of the Windows Performance Counter Object Name (e.g. Processor, Processor Information, Memory, etc) in the tag list.  According to the [Hardware Sizing Guidelines](https://docs.influxdata.com/influxdb/v0.13/guides/hardware_sizing/#when-do-i-need-more-ram), series cardinality strongly affects the amount of RAM required by the InfluxDB server.  Therefore, there is a risk that heavily instrumented Windows machines can unduly impact the provisioning requirements of the InfluxDB server simply due to the use of `win_perf_counters`.

The `wpc` plugin mitigates these two potential issues by making Performance Counter queries field names explicit, and by transparently regrouping fully-qualified Performance Counter queries by instance to minimize the number of points generated.

## An IIS / ASP.NET example

The templates below could be one way of instrumenting an IIS/ASP.NET web server.

```
  [[inputs.wpc.template]]
    Counters = [
      [ "get_rate", "\\Web Service(*)\\Get Requests/sec" ],
      [ "post_rate", "\\Web Service(*)\\Post Requests/sec" ],
      [ "conn_rate", "\\Web Service(*)\\Connection Attempts/sec" ],
      [ "conn", "\\Web Service(*)\\Current Connections" ],
      [ "isapi_rate", "\\Web Service(*)\\ISAPI Extension Requests/sec" ],
      # These queries are remapped to equivalent HAProxy field names.
      [ "qcur", "\\HTTP Service Request Queues(*)\\CurrentQueueSize" ],
      [ "dreq", "\\HTTP Service Request Queues(*)\\RejectedRequests" ],
    ]
    Measurement = "iis_websvc"
    ## Example output
    #  iis_websvc,instance=HelloWorld get_rate=4i,post_rate=10i,conn_rate=1i,conn=100i,isapi_rate=0i,qcur=0i,dreq=0i 1462765437090957980

  [[inputs.wpc.template]]
    Counters = [
      [ "restart", "\\ASP.NET\\Application Restarts" ],
      [ "wait_time", "\\ASP.NET\\Request Wait Time" ],
      [ "requests", "\\ASP.NET\\Requests Current" ],
      [ "waiting", "\\ASP.NET\\Requests Queued" ],
      [ "rejected", "\\ASP.NET\\Requests Rejected" ],
      [ "cache_tot", "\\ASP.NET Applications(__Total__)\\Cache Total Entries" ],
      [ "cache_hit", "\\ASP.NET Applications(__Total__)\\Cache Total Hit Ratio" ],
      [ "error_rate", "\\ASP.NET Applications(__Total__)\\Errors Total/Sec" ],
      [ "req_rate", "\\ASP.NET Applications(__Total__)\\Requests/Sec" ],
      [ "user_hit_rate", "\\Web Service Cache\\URI Cache Hits %" ],
      [ "system_hit_rate", "\\Web Service Cache\\Kernel: URI Cache Hits %" ],
      [ "file_hit_rate", "\\Web Service Cache\\File Cache Hits %" ],
    ]
    Measurement = "iis"
    ## Example output
    # iis,instance= restart=0i,wait_time=0i,requests=0i,waiting=0i,rejected=3i,user_hit_rate=4.99,system_hit_rate=3.34,file_hit_rate=9.3 1462765437090957980
    # iis,instance=__Total__ cache_tot=100000i,cache_hit=55.555555,error_rate=3.324,req_rate=0
```


## Exploring Windows Performance Counters with `typeperf`

Windows ships with a command-line utility named `typeperf` that can be used to query, explore, and sample the same fully-qualified Performance Counter queries that the `wpc` plugin expects.

```Dos
C:\> typeperf /?

Microsoft r TypePerf.exe (6.3.9600.17415)

Typeperf writes performance data to the command window or to a log file. To
stop Typeperf, press CTRL+C.

Usage:
typeperf { <counter [counter ...]> | -cf <filename> | -q [object]
                                | -qx [object] } [options]

Parameters:
  <counter [counter ...]>       Performance counters to monitor.

Options:
  -?                            Displays context sensitive help.
  -f <CSV|TSV|BIN|SQL>          Output file format. Default is CSV.
  -cf <filename>                File containing performance counters to
                                monitor, one per line.
  -si <[[hh:]mm:]ss>            Time between samples. Default is 1 second.
  -o <filename>                 Path of output file or SQL database. Default
                                is STDOUT.
  -q [object]                   List installed counters (no instances). To
                                list counters for one object, include the
                                object name, such as Processor.
  -qx [object]                  List installed counters with instances. To
                                list counters for one object, include the
                                object name, such as Processor.
  -sc <samples>                 Number of samples to collect. Default is to
                                sample until CTRL+C.
  -config <filename>            Settings file containing command options.
  -s <computer_name>            Server to monitor if no server is specified
                                in the counter path.
  -y                            Answer yes to all questions without prompting.

Note:
  Counter is the full name of a performance counter in
  "\\<Computer>\<Object>(<Instance>)\<Counter>" format,
  such as "\\Server1\Processor(0)\% User Time".

Examples:
  typeperf "\Processor(_Total)\% Processor Time"
  typeperf -cf counters.txt -si 5 -sc 50 -f TSV -o domain2.tsv
  typeperf -qx PhysicalDisk -o counters.txt
```


## Future Features

- Add the ability to alias Instance values and/or the ability to specify sets of equivalent instances (e.g. [ "", "_Total", "__Total__" ]).

