#New Relic output plugin
  
This plugins writes to New Relic insights.

```
[[outputs.newrelic]]
## New Relic Insights API key
insights_key = "insights api key"
##event_prefix if defined, prefix's metrics name for easy identification
# event_prefix = "Telegraf_"
```
####Parameters

|Parameter Name|Type|Description|
|:-|:-|:-|
| insights_key | Required | Insights API Insert key  |
| event_prefix | Optional | If defined, prefix's metrics name for easy identification |
