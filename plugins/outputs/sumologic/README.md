# SumoLogic Graphite Output Plugin

This plugin sends graphite metrics to [SumoLogic](https://help.sumologic.com/Metrics/Working-with-Metrics)
via http.

## Configuration:

```
# Configuration for SumoLogic server to send metrics to
[[outputs.sumologic]]
   ## Prefix metrics name
   prefix = "sumo-telegraf"
   ## Graphite output template
   template = "host.tags.measurement.field"
   ## Connection timeout.
   # timeout = "5s"
   ## SumoLogic Collector Url
   CollectorUrl = "<YOUR Sumo Logic collector url>" # required.
```
