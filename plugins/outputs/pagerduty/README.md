# PagerDuty output plugin

This plugin is used to send PagerDuty alerts based on a metric field.

Following is an example of send PagerDuty alerts based on "time_iowait" field
of the "cpu" metric. It will send PagerDyty alert any time the "time_iowait"
value is more than 50

Optionally, alerts can be restriced to metrics with a given set of tags

```toml
[[outputs.pagerduty]]
  service_key = "<SERVICE KEY>"
  metric = "cpu"
  description = "Check CPU"
  field = "time_iowait"
  expression = "> 50.0"
  [[outputs.pagerduty.tags]]
    name = "role"
    value = "web"
```
