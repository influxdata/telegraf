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
  # formats allowed on index_name after a prefix: 
  # %Y - year (2016)
  # %y - last two digits of year (00..99)
  # %m - month (01..12)
  # %d - day of month (e.g., 01)
  # %H - hour (00..23)
  index_name = "test"
  ## ElasticSearch uses a sniffing process to find all nodes of your cluster by default, automatically
  enable_sniffer = false
  ## Enable health check
  health_check = false
  ## If index not exists, a template will be created and then the new index.
  ##  You can set number of shards and replicas for this template.
  ## If the index's name uses formats ("myindex%Y%m%d"), the template's name will be the characters before
  ## the first '%' ("myindex").
  number_of_shards = 1
  number_of_replicas = 0
  
