# Solr input plugin

The [solr](http://lucene.apache.org/solr/) plugin collects stats via the 
[MBean Request Handler](https://cwiki.apache.org/confluence/display/solr/MBean+Request+Handler)

More about [performance statistics](https://cwiki.apache.org/confluence/display/solr/Performance+Statistics+Reference)

### Configuration:

```
[[inputs.solr]]
  ## specify a list of one or more Solr servers
  servers = ["http://localhost:8983"]

  ## specify a list of one or more Solr cores
  cores = ["main"]
```
