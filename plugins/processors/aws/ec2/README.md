# EC2 Metadata Processor Plugin

The EC2 Metadata processor plugin appends additional metadata from AWS to metrics associated with EC2 instances.

### Configuration:

```toml
## Annotate metrics from the cloudwatch plugin
[[processors.aws_metadata_ec2]]

  ## Specify the Amazon Region to operate in
  region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  # access_key = ""
  # secret_key = ""
  # token = ""
  # role_arn = ""
  # profile = ""
  # shared_credential_file = ""

  ## Specify the TTL for metadata lookups
  # cache_ttl = "1h"

  ## Process metrics from a Cloudwatch input plugin configured for the AWS/EC2 namespace
  ## Default is "cloudwatch_aws_ec2"
  # metric_names = ["cloudwatch_aws_ec2"]

  ## Metric tag that contains the EC2 Instance ID
  # id = "instance_id"

  ## Annotate metrics with EC2 Tags
  # tags = [ "Name", "ami_id", "instance_type"]
```

### Tags:

The plugin applies the following tags to metrics when configured:

* `instance_type` - the EC2 Instance Type for the instance
* `ami_id` - the AMI ID used by the instance
* Tags - for each configured tag name in the plugin, appends a tag if the EC2 Instance has a corresponding tag of that name
