# Elasticsearch Output Plugin

This plugin writes to [Elasticsearch](https://www.elastic.co) via Elastic (http://olivere.github.io/elastic/).

Tested with: 5.0.0-beta1 and 2.4.0

### Configuration:

```toml
# Configuration for Elasticsearch to send metrics to
[[outputs.elasticsearch]]
  ## The full HTTP endpoint URL for your Elasticsearch. # required
  server_host = "http://10.10.10.10:19200"
  ## The target index for metrics # required
  index_name = "twitter"
  ## ElasticSearch uses a sniffing process to find all nodes of your cluster by default, automatically
  enable_sniffer = false
  ## Earlier versions of EL doesn't accept "." in field name. Set delimiter with the character that you want instead.
  delimiter = "_"

