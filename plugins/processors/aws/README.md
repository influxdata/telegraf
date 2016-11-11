# AWS Metadata Processor Plugins

A series of plugins that extract additional metadata from AWS to annotate metrics.

### Configuration:

Each processor is scoped to a single AWS service (e.g. EC2, ELB, etc).

The processor plugins each share a common configuration pattern for 
configuring AWS credentials and basic processing information.

```toml
## Common AWS credential configuration

## Amazon Region (required)
region = "us-east-1"

## Amazon Credentials
## Credentials are loaded in the following order
## 1) Assumed credentials via STS if role_arn is specified
## 2) explicit credentials from 'access_key' and 'secret_key'
## 3) shared profile from 'profile'
## 4) environment variables
## 5) shared credentials file
## 6) EC2 Instance Profile
#access_key = ""
#secret_key = ""
#token = ""
#role_arn = ""
#profile = ""
#shared_credential_file = ""

## Common processing configuration

## Specify the TTL for metadata lookups
#cache_ttl = "1h"

## Specify the metric names to annotate with this processor
## By default the processor is configured to process the associated metric name from the Cloudwatch input plugin
#metric_names = [ "cloudwatch_aws_ec2" ]

## Specify the metric tag which contains the AWS resource ID
## By default the plugin is configured to find the resource's ID in the tag created by the Cloudwatch input plugin
#id = "instance_id"

## Plugin specific configuration
## Configure specific annotations available for this processor
```

### Processor Plugins:

* [aws_metadata_ec2](./ec2)
* [aws_metadata_elb](./elb)
* [aws_metadata_rds](./rds)
* [aws_metadata_sqs](./sqs)
