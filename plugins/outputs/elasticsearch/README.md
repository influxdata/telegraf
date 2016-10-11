# Elasticsearch Output Plugin

This plugin writes to [Elasticsearch](https://www.elastic.co) via Elastic (http://olivere.github.io/elastic/).

Attention: 
Elasticsearch 2.x does not support this dots-to-object transformation and so dots in field names are not allowed in versions 2.X.
In this case, dots will be replaced with "_".

### Configuration:

```toml
# Configuration for Elasticsearch to send metrics to
[[outputs.elasticsearch]]
  ## The full HTTP endpoint URL for your Elasticsearch. # required
  server_host = "http://10.10.10.10:9200"
  ## The target index for metrics # required
  index_name = "test"
  ## ElasticSearch uses a sniffing process to find all nodes of your cluster by default, automatically
  enable_sniffer = false
