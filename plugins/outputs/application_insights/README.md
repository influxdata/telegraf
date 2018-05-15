# Application Insights Output Plugin

This plugin writes telegraf metrics to Azure Application Insights

## Configuration
```
[[outputs.application_insights]]
  ## Instrumentation key of the Application Insights resource.
  instrumentationKey = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"

  ## Timeout on close. If not provided, will default to 5s. 0s means no timeout (not recommended).
  # timeout = "5s"

  ## Determines whether diagnostic logging (for Application Insights endpoint traffic) is enabled. Default is false.
  # enable_diagnosic_logging = "true"

  ## ContextTagSources dictionary instructs the Application Insights plugin to set Application Insights context tags using metric properties.
  ## In this dictionary keys are Application Insights context tags to set, and values are names of metric properties to use as source of data.
  ## For example:
  # [outputs.application_insights.context_tag_sources]
  # "ai.cloud.role" = "kubernetes_container_name"
  # "ai.cloud.roleInstance" = "kubernetes_pod_name"
  ## will set the ai.cloud.role context tag to the value of kubernetes_container_name property (if present), 
  ## and the ai.cloud.roleInstance context tag to the value of kubernetes_pod_name property.
  ## For list of all context tag keys see https://github.com/Microsoft/ApplicationInsights-Go/blob/master/appinsights/contracts/contexttagkeys.go
```

## Implementation notes
- Every field in a metric will result in a separate metric telemetry. For example, the metric `foo,host=a first=42,second=43 1525293034000000000`
will result in two metric telemetry records sent to Application Insights: first named `foo_first` and value of 42, and the secod named `foo_second` and a value of 43 (both having property `host` set to "a". \
\
The exception is a single-field metric with a value named `value`, in that case the single metric telemetry created will use just the whole metric name without the "value" suffix. For example, `bar,host=a value=23 1525293034000000000` will result in a telemetry named `bar` and value 23.