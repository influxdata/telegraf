# Application Insights Output Plugin

This plugin writes telegraf metrics to [Azure Application Insights](https://azure.microsoft.com/en-us/services/application-insights/).

### Configuration:
```toml
[[outputs.application_insights]]
  ## Instrumentation key of the Application Insights resource.
  instrumentation_key = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"

  ## Timeout for closing (default: 5s).
  # timeout = "5s"

  ## Enable additional diagnostic logging.
  # enable_diagnostic_logging = false

  ## Context Tag Sources add Application Insights context tags to a tag value.
  ##
  ## For list of allowed context tag keys see:
  ## https://github.com/Microsoft/ApplicationInsights-Go/blob/master/appinsights/contracts/contexttagkeys.go
  # [outputs.application_insights.context_tag_sources]
  #   "ai.cloud.role" = "kubernetes_container_name"
  #   "ai.cloud.roleInstance" = "kubernetes_pod_name"
```


### Metric Encoding:

For each field an Application Insights Telemetry record is created named based
on the measurement name and field.


**Example:** Create the telemetry records `foo_first` and `foo_second`:
```
foo,host=a first=42,second=43 1525293034000000000
```

In the special case of a single field named `value`, a single telemetry record is created named using only the measurement name

**Example:** Create a telemetry record `bar`:
```
bar,host=a value=42 1525293034000000000
```
