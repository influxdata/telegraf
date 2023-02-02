# OpenSearch Query Input Plugin

This [OpenSearch](https://opensearch.org/) plugin queries endpoints
to derive metrics from data stored in an OpenSearch cluster.

The following is supported:

- return number of hits for a search query
- calculate the `avg`/`max`/`min`/`sum` for a numeric field, filtered by a query,
  aggregated per tag
- `value_count` returns the number of documents for a particular field
- `stats` (returns `sum`, `min`, `max`, `avg`, and `value_count` in one query)
- extended_stats (`stats` plus stats such as sum of squares, variance, and standard
  deviation)
- `percentiles` returns the 1st, 5th, 25th, 50th, 75th, 95th, and 99th percentiles

## OpenSearch Support

This plugins is tested against OpenSearch 2.5.0 and 1.3.7.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Derive metrics from aggregating OpenSearch query results
[[inputs.opensearch_query]]
  ## OpenSearch cluster endpoint(s). Multiple urls can be specified as part
  ## of the same cluster.  Only one succesful call will be made per interval.
  urls = [ "https://node1.os.example.com:9200" ] # required.

  ## OpenSearch client timeout, defaults to "5s".
  # timeout = "5s"

  ## HTTP basic authentication details
  # username = "admin"
  # password = "admin"

  ## Skip TLS validation.  Useful for local testing and self-signed certs.
  # insecure_skip_verify = false

  [[inputs.opensearch_query.aggregation]]
    ## measurement name for the results of the aggregation query
    measurement_name = "measurement"

    ## OpenSearch index or index pattern to search
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

    ## Fields to be used as tags.  Must be text, non-analyzed fields. Metric
    ## aggregations are performed per tag
    # tags = ["field.keyword", "field2.keyword"]

    ## Set to true to not ignore documents when the tag(s) above are missing
    # include_missing_tag = false

    ## String value of the tag when the tag does not exist
    ## Required when include_missing_tag is true
    # missing_tag_value = "null"
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
  "min", "max", "sum", "value_count", "stats", "extended_stats", "percentiles".
  (see the [aggregation docs][agg]
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

### Example configurations

#### Search the average response time, per URI and per response status code

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

#### Search the maximum response time per method and per URI

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

#### Search number of documents matching a filter query in all indices

```toml
[[inputs.opensearch_query.aggregation]]
  measurement_name = "http_logs"
  index = "*"
  filter_query = "product_1 AND HEAD"
  query_period = "1m"
  date_field = "@timestamp"
```

#### Search number of documents matching a filter query, returning per response status code

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

#### Search all documents and generate common statistics, returning per response status code

```toml
[[inputs.opensearch_query.aggregation]]
  measurement_name = "http_logs"
  index = "*"
  tags = ["response.keyword"]
  include_missing_tag = false
  date_field = "@timestamp"
  query_period = "1m"
```

## Metrics

All metrics derive from aggregating OpenSearch query results.  Queries must
conform to appropriate OpenSearch
[Aggregations](https://opensearch.org/docs/latest/opensearch/aggregations/)
for more information.

Metric names are composed of a combination of the field name, metric aggregation
function, and the result field name.

For simple metrics, the result field name is `value`, and so getting the `avg`
on a field named `size` would produce the result `size_value_avg`.

For functions with multiple metrics, we use the resulting field.  For example,
the `stats` function returns five different results, so for a field `size`,
we would see five metric fields, named `size_stats_min`,
`size_stats_max`, `size_stats_sum`, `size_stats_avg`, and `size_stats_count`.

Nested results will build on their parent field names, for example, results for
percentile take the form:

```json
{
  "aggregations" : {
  "size_percentiles" : {
    "values" : {
      "1.0" : 21.984375,
      "5.0" : 27.984375,
      "25.0" : 44.96875,
      "50.0" : 64.22061688311689,
      "75.0" : 93.0,
      "95.0" : 156.0,
      "99.0" : 222.0
    }
  }
 }
}
```

Thus, our results would take the form `size_percentiles_values_1.0`.  This
structure applies to `percentiles` and `extended_stats` functions.

Note: `extended_stats` is currently limited to 2 standard deviations only.

## Example Output

```toml
[[inputs.opensearch_query.aggregation]]
    measurement_name = "bytes_stats"
    index = "opensearch_dashboards_sample_data_logs"
    date_field = "timestamp"
    query_period = "10m"
    filter_query = "*"
    metric_fields = ["bytes"]
    metric_function = "stats"
    tags = ["response.keyword"]
```

```text
bytes_stats,host=localhost,response_keyword=200 bytes_stats_sum=22231,doc_count=4i,bytes_stats_count=4,bytes_stats_min=941,bytes_stats_max=9544,bytes_stats_avg=5557.75 1672327840000000000
bytes_stats,host=localhost,response_keyword=404 bytes_stats_min=5330,bytes_stats_max=5330,bytes_stats_avg=5330,doc_count=1i,bytes_stats_sum=5330,bytes_stats_count=1 1672327840000000000
```
