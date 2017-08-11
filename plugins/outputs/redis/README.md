# Redis Output Plugin

This plug-in write the metrics collected by Telegraf in the REDIS LIST structure in the specified format.

### Configuration:

```
[[outputs.redis]]
#   ## redis service listen addr:port, default 127.0.0.1
    server_addr = "127.0.0.1:6379"
#   ## redis service login password (empty string does not execute the AUTH command)
#   # server_passwd = ""
#   ## redis list name, defalut telegraf/output
    queue_name = "telegraf/output"
#   ## Data format to output.
#   ## Each data format has its own unique set of configuration options, read
#   ## more about them here:
#   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
    data_format = "json"
```
