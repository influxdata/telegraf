# ELB Metadata Processor Plugin

The ELB Metadata processor plugin appends additional metadata from AWS to metrics associated with Elastic Load Balancers.

### Configuration:

```toml
## Annotate metrics from the cloudwatch plugin
[[processors.aws_metadata_elb]]

## Specify the Amazon Region to operate in
region = "us-east-1"

## Specify the TTL for metadata lookups
#cache_ttl = "1h"

## Process metrics from a Cloudwatch input plugin configured for the AWS/ELB namespace
## Default is "cloudwatch_aws_elb"
#metric_names = ["cloudwatch_aws_elb"]

## Metric tag that contains the Load Balancer Name
#id = "load_balancer_name"

## Annotate metrics with ELB Tags
#tags = [ "Name" ]
```

### Tags:

The plugin applies the following tags to metrics when configured:

* Tags - for each configured tag name in the plugin, appends a tag if the ELB has a corresponding tag of that name
