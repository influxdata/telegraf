# RDS Metadata Processor Plugin

The RDS Metadata processor plugin appends additional metadata from AWS to metrics associated with RDS instances.

### Configuration:

```toml
## Annotate metrics from the cloudwatch plugin
[[processors.aws_metadata_rds]]

## Specify the Amazon Region to operate in
region = "us-east-1"

## Specify the TTL for metadata lookups
#cache_ttl = "1h"

## Process metrics from a Cloudwatch input plugin configured for the AWS/RDS namespace
## Default is "cloudwatch_aws_rds"
#metric_names = ["cloudwatch_aws_rds"]

## Metric tag that contains the RDS DB Instance Identifier
#id = "db_instance_identifier"

## Annotate metrics with the RDS engine type
#engine = true

## Annotate metrics with the engine version
#engine_version = true

## Annotate metrics with RDS Tags
#tags = [ "Name" ]
```

### Tags:

The plugin applies the following tags to metrics when configured:

* `engine` - the RDS Engine type for the DB instance
* `engine_Version` - the RDS engine version for the DB Instance
* Tags - for each configured tag name in the plugin, appends a tag if the RDS Instance has a corresponding tag of that name
