# Tencent Cloud Cloud Monitor Input Plugin

This plugin will pull Metrics from Tencent Cloud Cloud Monitor (CM).

### Tencent Cloud Authentication

This plugin uses [Access Key](https://intl.cloud.tencent.com/document/product/598/34228) for Authentication with the Tencent Cloud API.

### Configuration:

```toml
[[inputs.tencentcloudcm]]
  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint = "tencentcloudapi.com"
  # endpoint = ""

  ## The default period for Tencent Cloud Cloud Monitor metrics is 1 minute (60s). However not all
  ## metrics are made available to the 1 minute period. Some are collected at
  ## 5 minute, 60 minute, or larger intervals.
  ## See: https://intl.cloud.tencent.com/document/product/248/33882
  ## Note that if a period is configured that is smaller than the default for a
  ## particular metric, that metric will not be returned by the Tencent Cloud API
  ## and will not be collected by Telegraf.
  ##
  ## Requested Tencent Cloud Cloud Monitor aggregation Period (required - must be a multiple of 60s)
  ## period = "5m"

  ## Collection Delay (must account for metrics availability via Tencent Cloud API)
  # delay = "0m"

  ## Maximum requests per second. Note that the global default Tencent Cloud API rate limit is
  ## 20 calls/second (1,200 calls/minute), so if you define multiple namespaces, these should add up to a
  ## maximum of 20.
  ## See https://intl.cloud.tencent.com/document/product/248/33881
  # ratelimit = 20

  ## Timeout for http requests made by the Tencent Cloud client.
  # timeout = "5s"

  ## By default, Tencent Cloud CM Input plugin will automatically discover instances in specified regions
  ## This sets the interval for discover and update the instances discovered.
  ##
  ## how often the discovery API call executed (default 1m)
  # discovery_interval = "1m"

  ## Tencent Cloud Account (required - you can provide multiple entries and distinguish them using
  ## optional name field, if name is empty, index number will be used as default)
  [[inputs.tencentcloudcm.accounts]]
    name = ""
    secret_id = ""
    secret_key = ""

    ## Namespaces to Pull
    [[inputs.tencentcloudcm.accounts.namespaces]]
      ## Tencent Cloud CM Namespace (required - see https://intl.cloud.tencent.com/document/product/248/34716#namespace)
      namespace = "QCE/CVM"

      ## Metrics filter, all metrics will be pulled if not left empty. Different namespaces may have different
      ## metric names, e.g. CVM Monitoring Metrics: https://intl.cloud.tencent.com/document/product/248/6843
      # metrics = ["CPUUsage", "MemUsage"]

      [[inputs.tencentcloudcm.accounts.namespaces.regions]]
        ## Tencent Cloud regions (required - Allowed values: https://intl.cloud.tencent.com/document/api/248/33876)
        region = "ap-guangzhou"

        ## Dimension filters for Metric. Different namespaces may have different
        ## dimension requirements, e.g. CVM Monitoring Metrics: https://intl.cloud.tencent.com/document/product/248/6843It must be specified if the namespace does not support instance auto discovery
        ## Currently, discovery supported for the following namespaces:
        ## - QCE/CVM
        ## - QCE/CDB
        ## - QCE/CES
        ## - QCE/REDIS
        ## - QCE/LB_PUBLIC
        ## - QCE/LB_PRIVATE
        ## - QCE/DC
        # [[inputs.tencentcloudcm.accounts.namespaces.regions.instances]]
        # [[inputs.tencentcloudcm.accounts.namespaces.regions.instances.dimensions]]
        #   name = "value"
```

#### Requirements and Terminology

Plugin Configuration utilizes [Tencent Cloud Cloud Monitor](https://intl.cloud.tencent.com/document/product/248/32799)

- `region` must be a valid Tencent Cloud [Region](https://intl.cloud.tencent.com/document/api/248/33876) value
- `period` must be a valid Tencent Cloud Cloud Monitor [Period](https://intl.cloud.tencent.com/document/product/248/33882) value
- `namespace` must be a valid Tencent Cloud Cloud Monitor [Namespace](https://intl.cloud.tencent.com/document/product/248/34716#namespace) value
- `names` must be valid Tencent Cloud Cloud Monitor [Metric](https://intl.cloud.tencent.com/document/product/248/34716#metric) names
- `dimensions` must be valid Tencent Cloud Cloud Monitor [Dimension](https://intl.cloud.tencent.com/document/product/248/34716#dimension) name/value pairs

### Measurements & Fields:

Each Tencent Cloud Cloud Monitor Namespace records a measurement with fields for each available Metric Statistic.
Namespace are represented in upper case

- {namespace}
  - value: (metric value)

### Tags;

Each measurement is tagged with the following identifiers:

- All measurements have the following tags:
  - account:          (Tencent Cloud Account)
  - metric:           (Tencent Cloud Cloud Monitor Metric)
  - namespace:        (Tencent Cloud Cloud Monitor Namespace)
  - period:           (Tencent Cloud Cloud Monitor Aggregation Period)
  - region:           (Tencent Cloud Region)
  - request_id:       (Tencent Cloud API Request ID)
  - {dimension-name}  (Tencent Cloud Dimension value)

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter tencentcloudcm --test
```
