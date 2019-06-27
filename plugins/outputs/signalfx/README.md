# SignalFx Output Plugin

```toml
    ## SignalFx API Token
    APIToken = "my-secret-key" # required.

    ## Ingest URL
    DatapointIngestURL = "https://ingest.signalfx.com/v2/datapoint"
    EventIngestURL = "https://ingest.signalfx.com/v2/event"

    ## Exclude metrics by metric name
    Exclude = ["plugin.metric_name", "plugin2.metric_name"]

    ## Events or String typed metrics are omitted by default,
    ## with the exception of host property events which are emitted by 
    ## the SignalFx Metadata Plugin.  If you require a string typed metric
    ## you must specify the metric name in the following list
    Include = ["plugin.metric_name", "plugin2.metric_name"]
```