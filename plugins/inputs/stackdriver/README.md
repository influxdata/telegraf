# Stackdriver Input Plugin

Stackdriver scrapes metrics from Google Cloud Monitoring.

### Credentials

This plugin uses a [Google Service Account](https://cloud.google.com/docs/authentication/getting-started) to interacting with the Stackdriver Monitoring API.

### Configuration:

```toml
[[inputs.stackdriver]]
  ## GCP Project (required - must be prefixed with "projects/")
  project = "projects/{project_id_or_number}"
  
  ## API rate limit. On a default project, it seems that a single user can make
  ## ~14 requests per second. This might be configurable. Each API request can
  ## fetch every time series for a single metric type, though -- this is plenty
  ## fast for scraping all the builtin metric types (and even a handful of
  ## custom ones) every 60s.
  # rateLimit = 14
  
  ## Collection Delay Seconds (required - must account for metrics availability via Stackdriver Monitoring API)
  # delaySeconds = 60

  ## The first query to stackdriver queries for data points that have timestamp t
  ## such that: (now() - delaySeconds - lookbackSeconds) <= t <= (now() - delaySeconds).
  ## The subsequence queries to stackdriver query for data points that have timestamp t
  ## such that: lastQueryEndTime <= t <= (now() - delaySeconds).
  ## Note that influx will de-dedupe points that are pulled twice,
  ## so it's best to be safe here, just in case it takes GCP awhile
  ## to get around to recording the data you seek.
  # lookbackSeconds = 120

  ## Metric collection period
  interval = "1m"
  
  ## Configure the TTL for the internal cache of timeseries requests.
  ## Defaults to 1 hr if not specified
  # cacheTTLSeconds = 3600
  
  ## Sets whether or not to scrape all bucket counts for metrics whose value
  ## type is "distribution". If those ~70 fields per metric
  ## type are annoying to you, try out the distributionAggregationAligners
  ## configuration option, wherein you may specifiy a list of aggregate functions
  ## (e.g., ALIGN_PERCENTILE_99) that might be more useful to you.
  # scrapeDistributionBuckets = true
  
  ## Excluded GCP metric types. Any string prefix works.
  ## Only declare either this or includeMetricTypePrefixes
  excludeMetricTypePrefixes = [
    "agent",
    "aws",
    "custom"
  ]
  
  ## *Only* include these GCP metric types. Any string prefix works
  ## Only declare either this or excludeMetricTypePrefixes
  # includeMetricTypePrefixes = nil
  
  ## Excluded GCP metric and resource tags. Any string prefix works.
  ## Only declare either this or includeTagPrefixes
  excludeTagPrefixes = [
    "pod_id",
  ]
  
  ## *Only* include these GCP metric and resource tags. Any string prefix works
  ## Only declare either this or excludeTagPrefixes
  # includeTagPrefixes = nil
  
  ## Declares a list of aggregate functions to be used for metric types whose
  ## value type is "distribution". These aggregate values are recorded in the
  ## distribution's measurement *in addition* to the bucket counts. That is to
  ## say: setting this option is not mutually exclusive with
  ## scrapeDistributionBuckets.
  distributionAggregationAligners = [
    "ALIGN_PERCENTILE_99",
    "ALIGN_PERCENTILE_95",
    "ALIGN_PERCENTILE_50",
  ]
  
  ## The filter string consists of logical AND of the 
  ## resource labels and metric labels if both of them
  ## are specified. (optional)
  ## See: https://cloud.google.com/monitoring/api/v3/filters
  ## Declares resource labels to filter GCP metrics
  ## that match any of them.
  # [[inputs.stackdriver.filter.resourceLabels]]
  #   key = "instance_name"
  #   value = 'starts_with("localhost")'
  
  ## Declares metric labels to filter GCP metrics
  ## that match any of them.
  # [[inputs.stackdriver.filter.metricLabels]]
  #   key = "device_name"
  #   value = 'one_of("sda", "sdb")'
```

### Tips

- If `includeMetricTypePrefixes` field is specified, this plugin will add filter string into list metric descriptors request which is more efficient.
- If `includeMetricTypePrefixes` field is left blank, this plugin will fetch all metric descriptors. Usually this will cause more API cost.
