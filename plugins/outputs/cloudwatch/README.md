## Amazon CloudWatch Output for Telegraf

This plugin will send metrics to Amazon CloudWatch.

## Amazon Authentication

This plugin uses a credential chain for Authentication with the CloudWatch
API endpoint. In the following order the plugin will attempt to authenticate.
1. Assumed credentials via STS if `role_arn` attribute is specified (source credentials are evaluated from subsequent rules)
2. Explicit credentials from `access_key`, `secret_key`, and `token` attributes
3. Shared profile from `profile` attribute
4. [Environment Variables](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables)
5. [Shared Credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)
6. [EC2 Instance Profile](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)

The IAM user needs only the `cloudwatch:PutMetricData` permission.

## Config

For this output plugin to function correctly the following variables
must be configured.

* region
* namespace

### region

The region is the Amazon region that you wish to connect to.
Examples include but are not limited to:
* us-west-1
* us-west-2
* us-east-1
* ap-southeast-1
* ap-southeast-2

### namespace

The namespace used for AWS CloudWatch metrics.

### write_statistics

If you have a large amount of metrics, you should consider to send statistic 
values instead of raw metrics which could not only improve performance but 
also save AWS API cost. If enable this flag, this plugin would parse the required 
[CloudWatch statistic fields](https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatch/#StatisticSet) 
(count, min, max, and sum) and send them to CloudWatch. You could use `basicstats` 
aggregator to calculate those fields. If not all statistic fields are available, 
all fields would still be sent as raw metrics.