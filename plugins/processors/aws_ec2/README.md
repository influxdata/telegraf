# AWS EC2 Metadata Processor Plugin

AWS EC2 Metadata processor plugin appends metadata gathered from [AWS IMDS][]
to metrics associated with EC2 instances.

[AWS IMDS]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

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

  ## cache_ttl determines how long each cached item will remain in the cache before
  ## it is removed and subsequently needs to be queried for from the AWS API. By
  ## default, no items are cached.
  # cache_ttl = "0s"

  ## tag_cache_size determines how many of the values which are found in imds_tags
  ## or ec2_tags will be kept in memory for faster lookup on successive processing
  ## of metrics. You may want to adjust this if you have excessively large numbers
  ## of tags on your EC2 instances, and you are using the ec2_tags field. This
  ## typically does not need to be changed when using the imds_tags field.
  # tag_cache_size = 1000

  ## log_cache_stats will emit a log line periodically to stdout with details of
  ## cache entries, hits, misses, and evacuations since the last time stats were
  ## emitted. This can be helpful in determining whether caching is being effective
  ## in your environment. Stats are emitted every 30 seconds. By default, this
  ## setting is disabled.
  # log_cache_stats = false
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

## Notes

We use a single cache because telegraf's `AddTag` function models this.

A user can specify a list of both EC2 tags and IMDS tags. The items in this list
can, technically, be the same. This will result in a situation where the EC2
Tag's value will override the IMDS tags value.

Though this is undesirable, it is unavoidable because the `AddTag` function does
not support this case.

You should avoid using IMDS tags as EC2 tags because the EC2 tags will always
"win" due to them being processed in this plugin *after* IMDS tags.
