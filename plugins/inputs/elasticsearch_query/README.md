# Elasticsearch query input plugin

This [elasticsearch](https://www.elastic.co/) query plugin queries endpoints to obtain metrics from data stored in an Elasticsearch cluster.

The following is supported:

- return number of hits for a search query
- calculate the avg/max/min/sum for a numeric field, filtered by a query, aggregated per tag
- count number of terms for a particular field

## Elasticsearch support

This plugins is tested against Elasticsearch 5.x and 6.x releases.
Currently it is known to break on 7.x or greater versions.

## Configuration

```toml
# Derive metrics from aggregating Elasticsearch query results
[[inputs.elasticsearch_query]]
  ## The full HTTP endpoint URL for your Elasticsearch instance
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  urls = [ "http://node1.es.example.com:9200" ] # required.

  ## Elasticsearch client timeout, defaults to "5s".
  # timeout = "5s"

  ## Set to true to ask Elasticsearch a list of all cluster nodes,
  ## thus it is not necessary to list all nodes in the urls config option
  # enable_sniffer = false

  ## Set the interval to check if the Elasticsearch nodes are available
  ## This option is only used if enable_sniffer is also set (0s to disable it)
  # health_check_interval = "10s"

  ## HTTP basic authentication details (eg. when using x-pack)
  # username = "telegraf"
  # password = "mypassword"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  [[inputs.elasticsearch_query.aggregation]]
    ## measurement name for the results of the aggregation query
    measurement_name = "measurement"

    ## Elasticsearch indexes to query (accept wildcards).
    index = "index-*"

    ## The date/time field in the Elasticsearch index (mandatory).
    date_field = "@timestamp"

    ## If the field used for the date/time field in Elasticsearch is also using
    ## a custom date/time format it may be required to provide the format to
    ## correctly parse the field.
    ##
    ## If using one of the built in elasticsearch formats this is not required.
    # date_field_custom_format = ""

    ## Time window to query (eg. "1m" to query documents from last minute).
    ## Normally should be set to same as collection interval
    query_period = "1m"

    ## Lucene query to filter results
    # filter_query = "*"

    ## Fields to aggregate values (must be numeric fields)
    # metric_fields = ["metric"]

    ## Aggregation function to use on the metric fields
    ## Must be set if 'metric_fields' is set
    ## Valid values are: avg, sum, min, max, sum
    # metric_function = "avg"

    ## Fields to be used as tags
    ## Must be text, non-analyzed fields. Metric aggregations are performed per tag
    # tags = ["field.keyword", "field2.keyword"]

    ## Set to true to not ignore documents when the tag(s) above are missing
    # include_missing_tag = false

    ## String value of the tag when the tag does not exist
    ## Used when include_missing_tag is true
    # missing_tag_value = "null"
```

## Examples

Please note that the `[[inputs.elasticsearch_query]]` is still required for all of the examples below.

### Search the average response time, per URI and per response status code

```toml
[[inputs.elasticsearch_query.aggregation]]
  measurement_name = "http_logs"
  index = "my-index-*"
  filter_query = "*"
  metric_fields = ["response_time"]
  metric_function = "avg"
  tags = ["URI.keyword", "response.keyword"]
  include_missing_tag = true
  missing_tag_value = "null"
  date_field = "@timestamp"
  query_period = "1m"
```

### Search the maximum response time per method and per URI

```toml
[[inputs.elasticsearch_query.aggregation]]
  measurement_name = "http_logs"
  index = "my-index-*"
  filter_query = "*"
  metric_fields = ["response_time"]
  metric_function = "max"
  tags = ["method.keyword","URI.keyword"]
  include_missing_tag = false
  missing_tag_value = "null"
  date_field = "@timestamp"
  query_period = "1m"
```

### Search number of documents matching a filter query in all indices

```toml
[[inputs.elasticsearch_query.aggregation]]
  measurement_name = "http_logs"
  index = "*"
  filter_query = "product_1 AND HEAD"
  query_period = "1m"
  date_field = "@timestamp"
```

### Search number of documents matching a filter query, returning per response status code

```toml
[[inputs.elasticsearch_query.aggregation]]
  measurement_name = "http_logs"
  index = "*"
  filter_query = "downloads"
  tags = ["response.keyword"]
  include_missing_tag = false
  date_field = "@timestamp"
  query_period = "1m"
```

### Required parameters

- `measurement_name`: The target measurement to be stored the results of the aggregation query.
- `index`: The index name to query on Elasticsearch
- `query_period`: The time window to query (eg. "1m" to query documents from last minute). Normally should be set to same as collection
- `date_field`: The date/time field in the Elasticsearch index

### Optional parameters

- `date_field_custom_format`: Not needed if using one of the built in date/time formats of Elasticsearch, but may be required if using a custom date/time format. The format syntax uses the [Joda date format](https://www.elastic.co/guide/en/elasticsearch/reference/6.8/search-aggregations-bucket-daterange-aggregation.html#date-format-pattern).
- `filter_query`: Lucene query to filter the results (default: "\*")
- `metric_fields`: The list of fields to perform metric aggregation (these must be indexed as numeric fields)
- `metric_funcion`: The single-value metric aggregation function to be performed on the `metric_fields` defined. Currently supported aggregations are "avg", "min", "max", "sum". (see [https://www.elastic.co/guide/en/elasticsearch/reference/current/search-aggregations-metrics.html](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-aggregations-metrics.html)
- `tags`: The list of fields to be used as tags (these must be indexed as non-analyzed fields). A "terms aggregation" will be done per tag defined
- `include_missing_tag`: Set to true to not ignore documents where the tag(s) specified above does not exist. (If false, documents without the specified tag field will be ignored in `doc_count` and in the metric aggregation)
- `missing_tag_value`: The value of the tag that will be set for documents in which the tag field does not exist. Only used when `include_missing_tag` is set to `true`.
