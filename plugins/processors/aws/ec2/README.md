# AWS EC2 Metadata Processor Plugin

AWS EC2 Metadata processor plugin appends metadata gathered from [AWS IMDS][]
to metrics associated with EC2 instances.

[AWS IMDS]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Attach AWS EC2 metadata to metrics
[[processors.aws_ec2]]
  ## Instance identity document tags to attach to metrics.
  ## For more information see:
  ## https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
  ##
  ## Available tags:
  ## * accountId
  ## * architecture
  ## * availabilityZone
  ## * billingProducts
  ## * imageId
  ## * instanceId
  ## * instanceType
  ## * kernelId
  ## * pendingTime
  ## * privateIp
  ## * ramdiskId
  ## * region
  ## * version
  imds_tags = []

  ## EC2 instance tags retrieved with DescribeTags action.
  ## In case tag is empty upon retrieval it's omitted when tagging metrics.
  ## Note that in order for this to work, role attached to EC2 instance or AWS
  ## credentials available from the environment must have a policy attached, that
  ## allows ec2:DescribeTags.
  ##
  ## For more information see:
  ## https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTags.html
  ec2_tags = []

  ## Timeout for http requests made by against aws ec2 metadata endpoint.
  timeout = "10s"

  ## ordered controls whether or not the metrics need to stay in the same order
  ## this plugin received them in. If false, this plugin will change the order
  ## with requests hitting cached results moving through immediately and not
  ## waiting on slower lookups. This may cause issues for you if you are
  ## depending on the order of metrics staying the same. If so, set this to true.
  ## Keeping the metrics ordered may be slightly slower.
  ordered = false

  ## max_parallel_calls is the maximum number of AWS API calls to be in flight
  ## at the same time.
  ## It's probably best to keep this number fairly low.
  max_parallel_calls = 10
```

## Example

Append `accountId` and `instanceId` to metrics tags:

```toml
[[processors.aws_ec2]]
  tags = [ "accountId", "instanceId"]
```

```diff
- cpu,hostname=localhost time_idle=42
+ cpu,hostname=localhost,accountId=123456789,instanceId=i-123456789123 time_idle=42
```
