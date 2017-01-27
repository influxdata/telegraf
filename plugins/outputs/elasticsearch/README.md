# Elasticsearch Output Plugin

This plugin writes to [Elasticsearch](https://www.elastic.co) via HTTP using Elastic (http://olivere.github.io/elastic/).

It only supports Elasticsearch 5.x series currently.

### Example:

The plugin will format the metrics in the following way:

```json
{
  "@timestamp": "2017-01-01T00:00:00+00:00",
  "input_plugin": "cpu",
  "cpu": {
    "usage_guest": 0,
    "usage_guest_nice": 0,
    "usage_idle": 71.85413456197966,
    "usage_iowait": 0.256805341656516,
    "usage_irq": 0,
    "usage_nice": 0,
    "usage_softirq": 0.2054442732579466,
    "usage_steal": 0,
    "usage_system": 15.04879301548127,
    "usage_user": 12.634822807288275
  },
  "tag": {
    "cpu": "cpu-total",
    "host": "elastichost",
    "dc": "datacenter1"
  }
}
```

```json
{
  "@timestamp": "2017-01-01T00:00:00+00:00",
  "input_plugin": "system",
  "system": {
    "load1": 0.78,
    "load15": 0.8,
    "load5": 0.8,
    "n_cpus": 2,
    "n_users": 2
  },
  "tag": {
    "host": "elastichost",
    "dc": "datacenter1"
  }
}
```

### Configuration:

```toml
# Configuration for Elasticsearch to send metrics to.
[[outputs.elasticsearch]]
  ## The full HTTP endpoint URL for your Elasticsearch instance
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  urls = [ "http://node1.es.example.com:9200" ] # required.
  ## Set to true to ask Elasticsearch a list of all cluster nodes,
  ## thus it is not necessary to list all nodes in the urls config option
  enable_sniffer = true
  ## Set the interval to check if the nodes are available, in seconds
  ## Setting to 0 will disable the health check (not recommended in production)
  health_check_interval = 10
  ## HTTP basic authentication details (eg. when using Shield)
  # username = "telegraf"
  # password = "mypassword"

  # Index Config
  ## The target index for metrics (telegraf will create it if not exists)
  ## You can use a different index per time frame using the patterns below
  # %Y - year (2016)
  # %y - last two digits of year (00..99)
  # %m - month (01..12)
  # %d - day of month (e.g., 01)
  # %H - hour (00..23)
  index_name = "telegraf-%Y.%m.%d" # required.

  ## Template Config
  ## set to true if you want telegraf to manage its index template
  manage_template = true
  ## The template name used for telegraf indexes
  template_name = "telegraf"
  ## set to true if you want to overwrite an existing template
  overwrite_template = false
```
