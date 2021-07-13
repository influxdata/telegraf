## Amazon CloudWatch Logs Output for Telegraf

This plugin will send logs to Amazon CloudWatch.

## Amazon Authentication

This plugin uses a credential chain for Authentication with the CloudWatch Logs
API endpoint. In the following order the plugin will attempt to authenticate.
1. Assumed credentials via STS if `role_arn` attribute is specified (source credentials are evaluated from subsequent rules)
2. Explicit credentials from `access_key`, `secret_key`, and `token` attributes
3. Shared profile from `profile` attribute
4. [Environment Variables](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables)
5. [Shared Credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)
6. [EC2 Instance Profile](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)

The IAM user needs the following permissions ( https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/permissions-reference-cwl.html):
- `logs:DescribeLogGroups` - required for check if configured log group exist	
- `logs:DescribeLogStreams` - required to view all log streams associated with a log group. 
- `logs:CreateLogStream` - required to create a new log stream in a log group.) 
- `logs:PutLogEvents` - required to upload a batch of log events into log stream. 

## Config
```toml
[[outputs.cloudwatch_logs]]
  ## The region is the Amazon region that you wish to connect to.
  ## Examples include but are not limited to:
  ## - us-west-1
  ## - us-west-2
  ## - us-east-1
  ## - ap-southeast-1
  ## - ap-southeast-2
  ## ...
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
  
  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Cloud watch log group. Must be created in AWS cloudwatch logs upfront!
  ## For example, you can specify the name of the k8s cluster here to group logs from all cluster in oine place
  log_group = "my-group-name" 
  
  ## Log stream in log group
  ## Either log group name or reference to metric attribute, from which it can be parsed:
  ## tag:<TAG_NAME> or field:<FIELD_NAME>. If log stream is not exist, it will be created.
  ## Since AWS is not automatically delete logs streams with expired logs entries (i.e. empty log stream) 
  ## you need to put in place appropriate house-keeping (https://forums.aws.amazon.com/thread.jspa?threadID=178855)
  log_stream = "tag:location"
  
  ## Source of log data - metric name
  ## specify the name of the metric, from which the log data should be retrieved.
  ## I.e., if you  are using docker_log plugin to stream logs from container, then
  ## specify log_data_metric_name  = "docker_log"
  log_data_metric_name  = "docker_log"
  
  ## Specify from which metric attribute the log data should be retrieved:
  ## tag:<TAG_NAME> or field:<FIELD_NAME>.
  ## I.e., if you  are using docker_log plugin to stream logs from container, then
  ## specify log_data_source  = "field:message" 
  log_data_source  = "field:message"
```