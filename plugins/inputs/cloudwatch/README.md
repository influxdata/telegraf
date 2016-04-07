# Amazon CloudWatch Statistics Input

This plugin will pull Metric Statistics from Amazon CloudWatch.

### Amazon Authentication

This plugin uses a credential chain for Authentication with the CloudWatch
API endpoint. In the following order the plugin will attempt to authenticate.
1. [IAMS Role](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)
2. [Environment Variables](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables)
3. [Shared Credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)

### Configuration:

```toml
[[inputs.cloudwatch]]
  ## Amazon Region (required)
  region = 'us-east-1'

  ## Requested CloudWatch aggregation Period (required - must be a multiple of 60s)
  period = '1m'

  ## Collection Delay (required - must account for metrics availability via CloudWatch API)
  delay = '1m'

  ## Override global run interval (optional - defaults to global interval)
  ## Recomended: use metric 'interval' that is a multiple of 'period' to avoid 
  ## gaps or overlap in pulled data
  interval = '1m'

  ## Metric Statistic Namespace (required)
  namespace = 'AWS/ELB'

  ## Metrics to Pull (optional)
  ## Defaults to all Metrics in Namespace if nothing is provided
  ## Refreshes Namespace available metrics every 1h
  [[inputs.cloudwatch.metrics]]
    names = ['Latency', 'RequestCount']
	
    ## Dimension filters for Metric (optional)
    [[inputs.cloudwatch.metrics.dimensions]]
      name = 'LoadBalancerName'
      value = 'p-example'
```
#### Requirements and Terminology

Plugin Configuration utilizes [CloudWatch concepts](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html) and access pattern to allow monitoring of any CloudWatch Metric.

- `region` must be a valid AWS [Region](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#CloudWatchRegions) value
- `period` must be a valid CloudWatch [Period](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#CloudWatchPeriods) value
- `namespace` must be a valid CloudWatch [Namespace](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Namespace) value
- `names` must be valid CloudWatch [Metric](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Metric) names
- `dimensions` must be valid CloudWatch [Dimension](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Dimension) name/value pairs

#### Restrictions and Limitations
- CloudWatch metrics are not available instantly via the CloudWatch API. You should adjust your collection `delay` to account for this lag in metrics availability based on your [monitoring subscription level](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-cloudwatch-new.html)
- CloudWatch API usage incurs cost - see [GetMetricStatistics Pricing](https://aws.amazon.com/cloudwatch/pricing/)

### Measurements & Fields:

Each CloudWatch Namespace monitored records a measurement with fields for each available Metric Statistic
Namespace and Metrics are represented in [snake case](https://en.wikipedia.org/wiki/Snake_case)

- cloudwatch_{namespace}
  - {metric}_sum         (metric Sum value)
  - {metric}_average     (metric Average value)
  - {metric}_minimum     (metric Minimum value)
  - {metric}_maximum     (metric Maximum value)
  - {metric}_sample_count (metric SampleCount value)


### Tags:
Each measurement is tagged with the following identifiers to uniquely identify the associated metric
Tag Dimension names are represented in [snake case](https://en.wikipedia.org/wiki/Snake_case)

- All measurements have the following tags:
  - region           (CloudWatch Region)
  - unit             (CloudWatch Metric Unit)
  - {dimension-name} (Cloudwatch Dimension value - one for each metric dimension)

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter cloudwatch -test
> cloudwatch_aws_elb,load_balancer_name=p-example,region=us-east-1,unit=seconds latency_average=0.004810798017284538,latency_maximum=0.1100282669067383,latency_minimum=0.0006084442138671875,latency_sample_count=4029,latency_sum=19.382705211639404 1459542420000000000
```
