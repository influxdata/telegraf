# Elasticsearch Output Plugin

This plugin writes to [Elasticsearch](https://www.elastic.co) via Elastic (http://olivere.github.io/elastic/).

### Configuration:

```toml
# Configuration for Elasticsearch to send metrics to
[[outputs.elasticsearch]]
  ## The full HTTP endpoint URL for your Elasticsearch.
  server_host = "http://10.10.10.10:19200"
  ## The target index for metrics # required
  index_name = "twitter"

