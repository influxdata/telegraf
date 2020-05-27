#New Relic output plugin
  
This plugins writes to New Relic insights.

```
[[outputs.newrelic]]
## New Relic Insights API key
insights_key = "insights api key"

# metric_prefix if defined, prefix's metrics name for easy identification
# metric_prefix = ""

# harvest timeout, default is 15 seconds
# timeout = "15s"
```
####Parameters

|Parameter Name|Type|Description|
|:-|:-|:-|
| insights_key | Required | Insights API Insert key  |
| metric_prefix | Optional | If defined, prefix's metrics name for easy identification |
| timeout | Optional | If defined, changes harvest timeout |
