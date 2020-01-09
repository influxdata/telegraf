# TDigest Aggregator Plugin

The tdigest aggregator plugin creates a histogram of each field it sees,
emitting the aggregations and the histogram every `period` seconds.

### Aggregation Concepts
* Tag key ```source``` has special meaning as the primary key of developers destination TSDB
* Tag key ```atom``` has special meaning to denote the atomic dimension of an aggregation
* Aggregation Macros
    * Macros are used to define the tags used as aggregation dimensions as well as what aggregate values are generated
    * Submitted as the value for the tag ```_rollup```
    * Only tags matching dimension list for aggregation.  Unused tags are removed from aggregate values
    * Format ```MACRO:Dim1;Dim2;...```
        * Wildcards ```*``` are supported 
            * Partial ```MACRO:Dim1;Othe*;Dim2```
            * Full ```MACRO:*``` : Uses every tag on the data for aggregation
    * Supported Macros
        - Timer
          - max,min,count,p99,p95,med
        - Counter
          - sum,count
        - Gauge
          - max,min,med,p95
        - Local
          - max,min,count,med
          - Does not generate histogram or keys for central aggregation

### Configuration:

```
# Keep the histogram of each metric passing through.
[[aggregators.tdigestagg]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "60s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = true

  ## TDigest Compression
  ## This value corresponds to the number of centroids the histogram will use
  ## Higher values increase size of data but also precision of calculated percentiles
  compression = 30.0

  ## CLAM: Cluster Level Aggregator for Metrics
  ## OSS Publication Pending - Insert link when available
  ## This is an Apache Spark job that was the original consumer of the output of this plugin
  ## The output format was considered too specialized for OSS Telegraf release but the functionality
  ##	has been preserved so that it can be leveraged at users discretion.
  using_clam = false

  ## One bucketing configuration is required to get any output
  [[aggregators.tdigestagg.bucketing]]
	## List of tags that will not be considered for aggregation and not emitted.
    ## Empty list is valid
	exclude_tags=[host]
	## Special Tag
    ## "source" is required in output by SLA but typically mapped from another input tag
    ## If source_tag_key is not set on an input point, a default value will be set and
    ## a sla_violation tag will be added
	source_tag_key=service
	## Optional: Default value is "atom"
    ## Special Tag
    ## "atom" is required for output by SLA.  Tag can be submitted with input points or mapped
    ## from another input tag.  If "atom" or configured replacement tag is not set on an input
    ## point, a default value will be set and a sla_violation tag will be added
	atom_replacement_tag_key=az

  ## Subsequent bucketing configurations will all ingest the same points
  [[aggregators.tdigest.bucketing]]
	exclude_tags=[]
	source_tag_key=service

  ## All supported macro names should be added here
  ## This logic could potentially be supported w/o config but this functionality already
  [aggregators.tdigestagg.tagpass]
    _rollup = ["timer*", "counter*", "gauge*", "local*", "default*"]         
```

### Measurements & Fields:

- measurement
  - centroids
  - compression

### Tags Added:

- aggregates
  - Comma separated list of aggregates to generate
- bucket_key
  - Unique key to use to combine histograms in central location
- source
  - Primary key for destination TSDB
- atom
  - Denotes the atomic dimension of an aggregation
  
### Tags Consumed:

- _rollup
  - Defines aggregation behavior

### Example Output:
TODO: Update sample output to show using_clam=false (default) behavior
```
Pretty printed for readability
{
  "fields": {
    "sum._utility": 1230.0,
    "centroids": "[{97.97979797979798 1} {97.97979797979798 1} {98 1} {98 1} {98 1} {98 1} {98 1} {98 1} {98.00990099009901 2} {98.01980198019803 2} {98.01980198019803 2} {98.01980198019803 2} {98.98989898989899 1} {98.98989898989899 2} {99 1} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99.00990099009901 2} {99.00990099009901 2} {99.00990099009901 2} {100 2} {100 2} {100 1} {100 1} {100 1} {100 1} {100 1} {100 1} {100 1}]",
    "compression": 30
    "using_clam": true
  },
  "name": "cpu_usage_idle",
  "tags": {
    "cpu": "cpu1",
    "source": "C02S121GG8WL.group.on",
    "az": "snc1",
    "env": "dev",
    "service": "awesome",
    "aggregates": "max,min,count,p99,p95,avg,med",
    "bucket_key": "cpu_usage_idle_awesome_snc1_dev"
  },
  "timestamp": 1532630290113371000
}
```
