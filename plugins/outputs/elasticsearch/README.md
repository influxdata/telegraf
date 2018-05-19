## Elasticsearch Output Plugin for Telegraf

This plugin writes to [Elasticsearch](https://www.elastic.co) via HTTP using Elastic (http://olivere.github.io/elastic/).

Currently it only supports Elasticsearch 5.x series.

## Elasticsearch indexes and templates

### Indexes per time-frame

This plugin can manage indexes per time-frame, as commonly done in other tools with Elasticsearch.

The timestamp of the metric collected will be used to decide the index destination.

For more information about this usage on Elasticsearch, check https://www.elastic.co/guide/en/elasticsearch/guide/master/time-based.html#index-per-timeframe

### Template management

Index templates are used in Elasticsearch to define settings and mappings for the indexes and how the fields should be analyzed.
For more information on how this works, see https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html

This plugin can create a working template for use with telegraf metrics. It uses Elasticsearch dynamic templates feature to set proper types for the tags and metrics fields.
If the template specified already exists, it will not overwrite unless you configure this plugin to do so. Thus you can customize this template after its creation if necessary.

Example of an index template created by telegraf:

```json
{
  "order": 0,
  "template": "telegraf-*",
  "settings": {
    "index": {
      "mapping": {
        "total_fields": {
          "limit": "5000"
        }
      },
      "refresh_interval": "10s"
    }
  },
  "mappings": {
    "_default_": {
      "dynamic_templates": [
        {
          "tags": {
            "path_match": "tag.*",
            "mapping": {
              "ignore_above": 512,
              "type": "keyword"
            },
            "match_mapping_type": "string"
          }
        },
        {
          "metrics_long": {
            "mapping": {
              "index": false,
              "type": "float"
            },
            "match_mapping_type": "long"
          }
        },
        {
          "metrics_double": {
            "mapping": {
              "index": false,
              "type": "float"
            },
            "match_mapping_type": "double"
          }
        },
        {
          "text_fields": {
            "mapping": {
              "norms": false
            },
            "match": "*"
          }
        }
      ],
      "_all": {
        "enabled": false
      },
      "properties": {
        "@timestamp": {
          "type": "date"
        },
        "measurement_name": {
          "type": "keyword"
        }
      }
    }
  },
  "aliases": {}
}

```

### Example events:

This plugin will format the events in the following way:

```json
{
  "@timestamp": "2017-01-01T00:00:00+00:00",
  "measurement_name": "cpu",
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
  "measurement_name": "system",
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
  ## Elasticsearch client timeout, defaults to "5s" if not set.
  timeout = "5s"
  ## Set to true to ask Elasticsearch a list of all cluster nodes,
  ## thus it is not necessary to list all nodes in the urls config option
  enable_sniffer = false
  ## Set the interval to check if the Elasticsearch nodes are available
  ## Setting to "0s" will disable the health check (not recommended in production)
  health_check_interval = "10s"
  ## HTTP basic authentication details (eg. when using Shield)
  # username = "telegraf"
  # password = "mypassword"

  ## Index Config
  ## The target index for metrics (Elasticsearch will create if it not exists).
  ## You can use the date specifiers below to create indexes per time frame.
  ## The metric timestamp will be used to decide the destination index name
  # %Y - year (2016)
  # %y - last two digits of year (00..99)
  # %m - month (01..12)
  # %d - day of month (e.g., 01)
  # %H - hour (00..23)
  # %V - week of the year (ISO week) (01..53)
  ## Additionally, you can specify a tag name using the notation {{tag_name}}
  ## which will be used as part of the index name. If the tag does not exist,
  ## the default tag value will be used.
  # index_name = "telegraf-{{host}}-%Y.%m.%d"
  # default_tag_value = "none"
  index_name = "telegraf-%Y.%m.%d" # required.

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Template Config
  ## Set to true if you want telegraf to manage its index template.
  ## If enabled it will create a recommended index template for telegraf indexes
  manage_template = true
  ## The template name used for telegraf indexes
  template_name = "telegraf"
  ## Set to true if you want telegraf to overwrite an existing template
  overwrite_template = false
```

### Required parameters:

* `urls`: A list containing the full HTTP URL of one or more nodes from your Elasticsearch instance.
* `index_name`: The target index for metrics. You can use the date specifiers below to create indexes per time frame.

```   %Y - year (2017)
  %y - last two digits of year (00..99)
  %m - month (01..12)
  %d - day of month (e.g., 01)
  %H - hour (00..23)
  %V - week of the year (ISO week) (01..53)
```
Additionally, you can specify dynamic index names by using tags with the notation ```{{tag_name}}```. This will store the metrics with different tag values in different indices. If the tag does not exist in a particular metric, the `default_tag_value` will be used instead.

### Optional parameters:

* `timeout`: Elasticsearch client timeout, defaults to "5s" if not set.
* `enable_sniffer`: Set to true to ask Elasticsearch a list of all cluster nodes, thus it is not necessary to list all nodes in the urls config option.
* `health_check_interval`: Set the interval to check if the nodes are available, in seconds. Setting to 0 will disable the health check (not recommended in production).
* `username`: The username for HTTP basic authentication details (eg. when using Shield).
* `password`: The password for HTTP basic authentication details (eg. when using Shield).
* `manage_template`: Set to true if you want telegraf to manage its index template. If enabled it will create a recommended index template for telegraf indexes.
* `template_name`: The template name used for telegraf indexes.
* `overwrite_template`: Set to true if you want telegraf to overwrite an existing template.

## Known issues

Integer values collected that are bigger than 2^63 and smaller than 1e21 (or in this exact same window of their negative counterparts) are encoded by golang JSON encoder in decimal format and that is not fully supported by Elasticsearch dynamic field mapping. This causes the metrics with such values to be dropped in case a field mapping has not been created yet on the telegraf index. If that's the case you will see an exception on Elasticsearch side like this:

```{"error":{"root_cause":[{"type":"mapper_parsing_exception","reason":"failed to parse"}],"type":"mapper_parsing_exception","reason":"failed to parse","caused_by":{"type":"illegal_state_exception","reason":"No matching token for number_type [BIG_INTEGER]"}},"status":400}```

The correct field mapping will be created on the telegraf index as soon as a supported JSON value is received by Elasticsearch, and subsequent insertions will work because the field mapping will already exist.

This issue is caused by the way Elasticsearch tries to detect integer fields, and by how golang encodes numbers in JSON. There is no clear workaround for this at the moment.
