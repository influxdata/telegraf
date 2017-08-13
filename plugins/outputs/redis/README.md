# Redis Output Plugin

This plug-in write the metrics collected by Telegraf in the REDIS LIST structure in the specified format.

### Configuration:

```
# [[outputs.redis]]
#   ## redis service listen addr:port, default 127.0.0.1
#   # server = "127.0.0.1:6379"
#   ## redis service login password
#   # password = ""
#   ## redis close connections after remaining idle for this duration.
#   ## if the value is zero, then idleconnections are not closed.
#   ## shoud set the timeout to a value lessthan the redis server's timeout.
#   # idle_timeout = "1s"
#   ## specifies the timeout for reading/writing a single command.
#   # timeout = "1s"
#   ## redis list name, defalut telegraf/output
#   # queue_name = "telegraf/output"
#   ## Data format to output.
#   ## Each data format has its own unique set of configuration options, read
#   ## more about them here:
#   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
#   # data_format = "influx"
```
