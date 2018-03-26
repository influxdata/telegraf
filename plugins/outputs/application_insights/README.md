# Application Insights Output Plugin

This plugin writes telegraf metrics to Auzre Application Insights

### Configuration
```
[[outputs.application_insights]]
  ## Instrumentation key of the Application Insights resource.
  instrumentationKey = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"

  ## Timeout on close. If not provided, will default to 5s. 0s means no timeout (not recommended).
  # timeout = "5s"
```
