# Amazon CloudWatch Output Plugin

This plugin will send metrics to Amazon CloudWatch.

## Amazon Authentication

This plugin uses a credential chain for Authentication with the CloudWatch API
endpoint. In the following order the plugin will attempt to authenticate.

1. Web identity provider credentials via STS if `role_arn` and
   `web_identity_token_file` are specified
1. Assumed credentials via STS if `role_arn` attribute is specified (source
   credentials are evaluated from subsequent rules)
1. Explicit credentials from `access_key`, `secret_key`, and `token` attributes
1. Shared profile from `profile` attribute
1. [Environment Variables][1]
1. [Shared Credentials][2]
1. [EC2 Instance Profile][3]

If you are using credentials from a web identity provider, you can specify the
session name using `role_session_name`. If left empty, the current timestamp
will be used.

The IAM user needs only the `cloudwatch:PutMetricData` permission.

[1]: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables
[2]: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file
[3]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html

## Configuration

```toml
# Configuration for AWS CloudWatch output.
[[outputs.cloudwatch]]
  ## Amazon REGION
  region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Web identity provider credentials via STS if role_arn and web_identity_token_file are specified
  ## 2) Assumed credentials via STS if role_arn is specified
  ## 3) explicit credentials from 'access_key' and 'secret_key'
  ## 4) shared profile from 'profile'
  ## 5) environment variables
  ## 6) shared credentials file
  ## 7) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #web_identity_token_file = ""
  #role_session_name = ""
  #profile = ""
  #shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Namespace for the CloudWatch MetricDatums
  namespace = "InfluxData/Telegraf"

  ## If you have a large amount of metrics, you should consider to send statistic
  ## values instead of raw metrics which could not only improve performance but
  ## also save AWS API cost. If enable this flag, this plugin would parse the required
  ## CloudWatch statistic fields (count, min, max, and sum) and send them to CloudWatch.
  ## You could use basicstats aggregator to calculate those fields. If not all statistic
  ## fields are available, all fields would still be sent as raw metrics.
  # write_statistics = false

  ## Enable high resolution metrics of 1 second (if not enabled, standard resolution are of 60 seconds precision)
  # high_resolution_metrics = false
```

For this output plugin to function correctly the following variables must be
configured.

* region
* namespace

### region

The region is the Amazon region that you wish to connect to.  Examples include
but are not limited to:

* us-west-1
* us-west-2
* us-east-1
* ap-southeast-1
* ap-southeast-2

### namespace

The namespace used for AWS CloudWatch metrics.

### write_statistics

If you have a large amount of metrics, you should consider to send statistic
values instead of raw metrics which could not only improve performance but also
save AWS API cost. If enable this flag, this plugin would parse the required
[CloudWatch statistic fields][1] (count, min, max, and sum) and send them to
CloudWatch. You could use `basicstats` aggregator to calculate those fields. If
not all statistic fields are available, all fields would still be sent as raw
metrics.

[1]: https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatch/#StatisticSet

### high_resolution_metrics

Enable high resolution metrics (1 second precision) instead of standard ones (60
seconds precision)
