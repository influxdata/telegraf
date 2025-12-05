# OpenSearch Output Plugin

This plugin writes metrics to a [OpenSearch][opensearch] instance via HTTP.
It supports OpenSearch releases v1 and v2 but future comparability with 1.x is
not guaranteed and instead will focus on 2.x support.

> [!TIP]
> Consider using the existing Elasticsearch plugin for 1.x.

⭐ Telegraf v1.29.0
🏷️ datastore, logging
💻 all

[opensearch]: https://opensearch.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Configuration for OpenSearch to send metrics to.
[[outputs.opensearch]]
  ## URLs
  ## The full HTTP endpoint URL for your OpenSearch instance. Multiple URLs can
  ## be specified as part of the same cluster, but only one URLs is used to
  ## write during each interval.
  urls = ["http://node1.os.example.com:9200"]

  ## Index Name
  ## Target index name for metrics (OpenSearch will create if it not exists).
  ## This is a Golang template (see https://pkg.go.dev/text/template)
  ## You can also specify
  ## metric name (`{{.Name}}`), tag value (`{{.Tag "tag_name"}}`), field value (`{{.Field "field_name"}}`)
  ## If the tag does not exist, the default tag value will be empty string "".
  ## the timestamp (`{{.Time.Format "xxxxxxxxx"}}`).
  ## For example: "telegraf-{{.Time.Format \"2006-01-02\"}}-{{.Tag \"host\"}}" would set it to telegraf-2023-07-27-HostName
  index_name = ""

  ## Timeout
  ## OpenSearch client timeout
  # timeout = "5s"

  ## Sniffer
  ## Set to true to ask OpenSearch a list of all cluster nodes,
  ## thus it is not necessary to list all nodes in the urls config option
  # enable_sniffer = false

  ## GZIP Compression
  ## Set to true to enable gzip compression
  # enable_gzip = false

  ## Health Check Interval
  ## Set the interval to check if the OpenSearch nodes are available
  ## Setting to "0s" will disable the health check (not recommended in production)
  # health_check_interval = "10s"

  ## Set the timeout for periodic health checks.
  # health_check_timeout = "1s"
  ## HTTP basic authentication details.
  # username = ""
  # password = ""
  ## HTTP bearer token authentication details
  # auth_bearer_token = ""

  ## Optional TLS Config
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable =
  ## Trusted root certificates for server
  # tls_ca = "/path/to/cafile"
  ## Used for TLS client certificate authentication
  # tls_cert = "/path/to/certfile"
  ## Used for TLS client certificate authentication
  # tls_key = "/path/to/keyfile"
  ## Send the specified TLS server name via SNI
  # tls_server_name = "kubernetes.example.com"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Template Config
  ## Manage templates
  ## Set to true if you want telegraf to manage its index template.
  ## If enabled it will create a recommended index template for telegraf indexes
  # manage_template = true

  ## Template Name
  ## The template name used for telegraf indexes
  # template_name = "telegraf"

  ## Overwrite Templates
  ## Set to true if you want telegraf to overwrite an existing template
  # overwrite_template = false

  ## Document ID
  ## If set to true a unique ID hash will be sent as
  ## sha256(concat(timestamp,measurement,series-hash)) string. It will enable
  ## data resend and update metric points avoiding duplicated metrics with
  ## different id's
  # force_document_id = false

  ## Value Handling
  ## Specifies the handling of NaN and Inf values.
  ## This option can have the following values:
  ##    none    -- do not modify field-values (default); will produce an error
  ##               if NaNs or infs are encountered
  ##    drop    -- drop fields containing NaNs or infs
  ##    replace -- replace with the value in "float_replacement_value" (default: 0.0)
  ##               NaNs and inf will be replaced with the given number, -inf with the negative of that number
  # float_handling = "none"
  # float_replacement_value = 0.0

  ## Pipeline Config
  ## To use a ingest pipeline, set this to the name of the pipeline you want to use.
  # use_pipeline = "my_pipeline"

  ## Pipeline Name
  ## Additionally, you can specify a tag name using the notation (`{{.Tag "tag_name"}}`)
  ## which will be used as the pipeline name (e.g. "{{.Tag "os_pipeline"}}").
  ## If the tag does not exist, the default pipeline will be used as the pipeline.
  ## If no default pipeline is set, no pipeline is used for the metric.
  # default_pipeline = ""
```

### Required parameters

* `urls`: A list containing the full HTTP URL of one or more nodes from your
  OpenSearch instance.
* `index_name`: The target index for metrics. You can use the date format

For example: "telegraf-{{.Time.Format \"2006-01-02\"}}" would set it to
"telegraf-2023-07-27". You can also specify metric name (`{{ .Name }}`), tag
value (`{{ .Tag \"tag_name\" }}`), and field value
(`{{ .Field \"field_name\" }}`).

If the tag does not exist, the default tag value will be empty string ""

## Permissions

If you are using authentication within your OpenSearch cluster, you need to
create an account and create a role with at least the manage role in the Cluster
Privileges category. Otherwise, your account will not be able to connect to your
OpenSearch cluster and send logs to your cluster.  After that, you need to
add "create_index" and "write" permission to your specific index pattern.

## OpenSearch indexes and templates

### Indexes per time-frame

This plugin can manage indexes per time-frame, as commonly done in other tools
with OpenSearch. The timestamp of the metric collected will be used to decide
the index destination. For more information about this usage on OpenSearch,
check [the docs][1].

[1]: https://opensearch.org/docs/latest/

### Template management

Index templates are used in OpenSearch to define settings and mappings for
the indexes and how the fields should be analyzed.  For more information on how
this works, see [the docs][2].

This plugin can create a working template for use with telegraf metrics. It uses
OpenSearch dynamic templates feature to set proper types for the tags and
metrics fields.  If the template specified already exists, it will not overwrite
unless you configure this plugin to do so. Thus you can customize this template
after its creation if necessary.

Example of an index template created by telegraf on OpenSearch 2.x:

```json
{
  "telegraf-2022.10.02" : {
    "aliases" : { },
    "mappings" : {
      "properties" : {
        "@timestamp" : {
          "type" : "date"
        },
        "disk" : {
          "properties" : {
            "free" : {
              "type" : "long"
            },
            "inodes_free" : {
              "type" : "long"
            },
            "inodes_total" : {
              "type" : "long"
            },
            "inodes_used" : {
              "type" : "long"
            },
            "total" : {
              "type" : "long"
            },
            "used" : {
              "type" : "long"
            },
            "used_percent" : {
              "type" : "float"
            }
          }
        },
        "measurement_name" : {
          "type" : "text",
          "fields" : {
            "keyword" : {
              "type" : "keyword",
              "ignore_above" : 256
            }
          }
        },
        "tag" : {
          "properties" : {
            "cpu" : {
              "type" : "text",
              "fields" : {
                "keyword" : {
                  "type" : "keyword",
                  "ignore_above" : 256
                }
              }
            },
            "device" : {
              "type" : "text",
              "fields" : {
                "keyword" : {
                  "type" : "keyword",
                  "ignore_above" : 256
                }
              }
            },
            "host" : {
              "type" : "text",
              "fields" : {
                "keyword" : {
                  "type" : "keyword",
                  "ignore_above" : 256
                }
              }
            },
            "mode" : {
              "type" : "text",
              "fields" : {
                "keyword" : {
                  "type" : "keyword",
                  "ignore_above" : 256
                }
              }
            },
            "path" : {
              "type" : "text",
              "fields" : {
                "keyword" : {
                  "type" : "keyword",
                  "ignore_above" : 256
                }
              }
            }
          }
        }
      }
    },
    "settings" : {
      "index" : {
        "creation_date" : "1664693522789",
        "number_of_shards" : "1",
        "number_of_replicas" : "1",
        "uuid" : "TYugdmvsQfmxjzbGRJ8FIw",
        "version" : {
          "created" : "136247827"
        },
        "provided_name" : "telegraf-2022.10.02"
      }
    }
  }
}

```

[2]: https://opensearch.org/docs/latest/opensearch/index-templates/

### Example events

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
    "host": "opensearhhost",
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
    "host": "opensearhhost",
    "dc": "datacenter1"
  }
}
```

## Known issues

Integer values collected that are bigger than 2^63 and smaller than 1e21 (or in
this exact same window of their negative counterparts) are encoded by golang
JSON encoder in decimal format and that is not fully supported by OpenSearch
dynamic field mapping. This causes the metrics with such values to be dropped in
case a field mapping has not been created yet on the telegraf index. If that's
the case you will see an exception on OpenSearch side like this:

```json
{
  "error": {
    "root_cause": [
      {"type": "mapper_parsing_exception", "reason": "failed to parse"}
    ],
    "type": "mapper_parsing_exception",
    "reason": "failed to parse",
    "caused_by": {
      "type": "illegal_state_exception",
      "reason": "No matching token for number_type [BIG_INTEGER]"
    }
  },
  "status": 400
}
```

The correct field mapping will be created on the telegraf index as soon as a
supported JSON value is received by OpenSearch, and subsequent insertions
will work because the field mapping will already exist.

This issue is caused by the way OpenSearch tries to detect integer fields,
and by how golang encodes numbers in JSON. There is no clear workaround for this
at the moment.
