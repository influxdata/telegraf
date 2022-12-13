# OpenSearch Query Input Plugin

This [opensearch_query](https://opensearch.org/) plugin queries endpoints
to obtain metrics from data stored in an OpenSearch cluster.

The following is supported:

- return number of hits for a search query
- calculate the avg/max/min/sum for a numeric field, filtered by a query,
  aggregated per tag
- count number of terms for a particular field

## OpenSearch Support

This plugins is tested against OpenSearch 2.4.0.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Derive metrics from aggregating OpenSearch query results
[[inputs.opensearch_query]]
  ## The full HTTP endpoint URL for your OpenSearch instance
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  urls = [ "http://node1.os.example.com:9200" ] # required.

  ## OpenSearch client timeout, defaults to "5s".
  # timeout = "5s"

  ## HTTP basic authentication details
  # username = "admin"
  # password = "admin"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  [[inputs.opensearch_query.aggregation]]
    ## measurement name for the results of the aggregation query
    measurement_name = "measurement"

    ## OpenSearch indexes to query (accept wildcards).
    index = "index-*"

    ## The date/time field in the OpenSearch index (mandatory).
    date_field = "@timestamp"

    ## If the field used for the date/time field in OpenSearch is also using
    ## a custom date/time format it may be required to provide the format to
    ## correctly parse the field.
    ##
    ## If using one of the built in OpenSearch formats this is not required.
    ## https://opensearch.org/docs/2.4/opensearch/supported-field-types/date/#built-in-formats
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
    ## Must be text, non-analyzed fields. Metric aggregations are performed
    ## per tag
    # tags = ["field.keyword", "field2.keyword"]

    ## Set to true to not ignore documents when the tag(s) above are missing
    # include_missing_tag = false

    ## String value of the tag when the tag does not exist
    ## Used when include_missing_tag is true
    # missing_tag_value = "null"
```

## Examples

Please note that the `[[inputs.opensearch_query]]` is still required for all
of the examples below.

### Search the average response time, per URI and per response status code

```toml
[[inputs.opensearch_query.aggregation]]
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
[[inputs.opensearch_query.aggregation]]
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
[[inputs.opensearch_query.aggregation]]
  measurement_name = "http_logs"
  index = "*"
  filter_query = "product_1 AND HEAD"
  query_period = "1m"
  date_field = "@timestamp"
```

### Search number of documents matching a filter query, returning per response status code

```toml
[[inputs.opensearch_query.aggregation]]
  measurement_name = "http_logs"
  index = "*"
  filter_query = "downloads"
  tags = ["response.keyword"]
  include_missing_tag = false
  date_field = "@timestamp"
  query_period = "1m"
```

### Required parameters

- `measurement_name`: The target measurement to be stored the results of the
  aggregation query.
- `index`: The index name to query on OpenSearch
- `query_period`: The time window to query (eg. "1m" to query documents from
  last minute). Normally should be set to same as collection
- `date_field`: The date/time field in the OpenSearch index

### Optional parameters

- `date_field_custom_format`: Not needed if using one of the built in date/time
  formats of OpenSearch, but may be required if using a custom date/time
  format. The format syntax uses the [Joda date format][joda].
- `filter_query`: Lucene query to filter the results (default: "\*")
- `metric_fields`: The list of fields to perform metric aggregation (these must
  be indexed as numeric fields)
- `metric_funcion`: The single-value metric aggregation function to be performed
  on the `metric_fields` defined. Currently supported aggregations are "avg",
  "min", "max", "sum". (see the [aggregation docs][agg]
- `tags`: The list of fields to be used as tags (these must be indexed as
  non-analyzed fields). A "terms aggregation" will be done per tag defined
- `include_missing_tag`: Set to true to not ignore documents where the tag(s)
  specified above does not exist. (If false, documents without the specified tag
  field will be ignored in `doc_count` and in the metric aggregation)
- `missing_tag_value`: The value of the tag that will be set for documents in
  which the tag field does not exist. Only used when `include_missing_tag` is
  set to `true`.

[joda]: https://opensearch.org/docs/2.4/opensearch/supported-field-types/date/#custom-formats
[agg]: https://opensearch.org/docs/2.4/opensearch/aggregations/

## Metrics

All metrics derive from aggregating OpenSearch query results.  Queries must
conform to appropriate OpenSearch
[Aggregations](https://opensearch.org/docs/latest/opensearch/aggregations/)
for more information.

## Example Output

```shell
./telegraf --config plugins/inputs/opensearch_query/dev/telegraf.conf --input-filter opensearch_query --test
2022-12-13T21:09:41Z I! Starting Telegraf 1.26.0-a96e9d38
2022-12-13T21:09:41Z I! Available plugins: 214 inputs, 9 aggregators, 26 processors, 21 parsers, 57 outputs, 2 secret-stores
2022-12-13T21:09:41Z I! Loaded inputs: opensearch_query
2022-12-13T21:09:41Z I! Loaded aggregators:
2022-12-13T21:09:41Z I! Loaded processors:
2022-12-13T21:09:41Z I! Loaded secretstores:
2022-12-13T21:09:41Z W! Outputs are not used in testing mode!
2022-12-13T21:09:41Z I! Tags enabled: host=localhost
```
