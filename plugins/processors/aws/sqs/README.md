# SQS Metadata Processor Plugin

The SQS Metadata processor plugin appends additional metadata from AWS to metrics associated with Simple Queue Service.

### Configuration:

```toml
## Annotate metrics from the cloudwatch plugin
[[processors.aws_metadata_sqs]]

## Specify the Amazon Region to operate in
region = "us-east-1"

## Specify the TTL for metadata lookups
#cache_ttl = "1h"

## Process metrics from a Cloudwatch input plugin configured for the AWS/SQS namespace
## Default is "cloudwatch_aws_sqs"
#metric_names = ["cloudwatch_aws_sqs"]

## Metric tag that contains the SQS queue name
#id = "queue_name"

```

### Tags:

The plugin applies the following tags to metrics when configured:

* (none) - there is currently no support metadata to extract for SQS queues.
