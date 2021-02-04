# AWS EC2 Metadata Processor Plugin

AWS EC2 Metadata processor plugin appends metadata gathered from [AWS IMDS][]
to metrics associated with EC2 instances.

[AWS IMDS]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html

## Configuration

```toml
[[processors.aws_ec2]]
  ## Tags to attach to metrics. Available tags:
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
  tags = []

  ## Timeout for http requests made by against AWS EC2 metadata endpoint.
  timeout = "10s"

  ## ordered controls whether or not the metrics need to stay in the same order
  ## this plugin received them in. If false, this plugin will change the order
  ## with requests hitting cached results moving through immediately and not
  ## waiting on slower lookups. This may cause issues for you if you are
  ## depending on the order of metrics staying the same. If so, set this to true.
  ## Keeping the metrics ordered may be slightly slower.
  ordered = false
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
